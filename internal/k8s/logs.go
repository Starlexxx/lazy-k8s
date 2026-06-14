package k8s

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
)

type LogOptions struct {
	Container    string
	Follow       bool
	TailLines    int64
	Previous     bool
	Timestamps   bool
	SinceSeconds *int64
}

type LogLine struct {
	Content string
	Error   error
}

func (c *Client) StreamPodLogs(
	ctx context.Context,
	namespace, podName string,
	opts LogOptions,
) (<-chan LogLine, error) {
	namespace = c.ns(namespace)

	podLogOpts := &corev1.PodLogOptions{
		Container:  opts.Container,
		Follow:     opts.Follow,
		Previous:   opts.Previous,
		Timestamps: opts.Timestamps,
	}

	if opts.TailLines > 0 {
		podLogOpts.TailLines = &opts.TailLines
	}

	if opts.SinceSeconds != nil {
		podLogOpts.SinceSeconds = opts.SinceSeconds
	}

	stream, err := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts).Stream(ctx)
	if err != nil {
		return nil, err
	}

	logChan := make(chan LogLine, 100)

	go pumpLogLines(ctx, stream, logChan)

	return logChan, nil
}

// pumpLogLines reads lines from stream into logChan until the stream ends or
// ctx is canceled. Sends select on ctx.Done() so the goroutine can't block
// forever on a full channel once the consumer is gone.
func pumpLogLines(ctx context.Context, stream io.ReadCloser, logChan chan<- LogLine) {
	defer close(logChan)
	defer func() { _ = stream.Close() }()

	reader := bufio.NewReader(stream)

	send := func(l LogLine) bool {
		select {
		case <-ctx.Done():
			return false
		case logChan <- l:
			return true
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					send(LogLine{Error: err})
				}

				return
			}

			if !send(LogLine{Content: line}) {
				return
			}
		}
	}
}

func (c *Client) GetPodLogSnapshot(
	ctx context.Context,
	namespace, podName string,
	opts LogOptions,
) (string, error) {
	namespace = c.ns(namespace)

	podLogOpts := &corev1.PodLogOptions{
		Container:  opts.Container,
		Previous:   opts.Previous,
		Timestamps: opts.Timestamps,
	}

	if opts.TailLines > 0 {
		podLogOpts.TailLines = &opts.TailLines
	}

	result := c.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts).Do(ctx)
	if result.Error() != nil {
		return "", result.Error()
	}

	raw, err := result.Raw()
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// GetPodsLogSnapshot fetches a log snapshot from every named pod concurrently
// and returns a single blob where each line is prefixed with "[podName] ".
// Pods that fail to return logs surface as error lines rather than aborting
// the whole call — one broken pod shouldn't hide logs from the rest.
func (c *Client) GetPodsLogSnapshot(
	ctx context.Context,
	namespace string,
	podNames []string,
	opts LogOptions,
) (string, error) {
	namespace = c.ns(namespace)

	if len(podNames) == 0 {
		return "", nil
	}

	type podResult struct {
		name    string
		content string
		err     error
	}

	results := make([]podResult, len(podNames))

	var wg sync.WaitGroup

	for i, name := range podNames {
		wg.Add(1)

		go func(idx int, podName string) {
			defer wg.Done()

			content, err := c.GetPodLogSnapshot(ctx, namespace, podName, opts)
			results[idx] = podResult{name: podName, content: content, err: err}
		}(i, name)
	}

	wg.Wait()

	var b strings.Builder

	for _, r := range results {
		if r.err != nil {
			b.WriteString("[" + r.name + "] <error: " + r.err.Error() + ">\n")

			continue
		}

		for line := range strings.SplitSeq(r.content, "\n") {
			if line == "" {
				continue
			}

			b.WriteString("[" + r.name + "] ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}
