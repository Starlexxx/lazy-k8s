package k8s

import (
	"bufio"
	"context"
	"io"

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

func (c *Client) StreamPodLogs(ctx context.Context, namespace, podName string, opts LogOptions) (<-chan LogLine, error) {
	if namespace == "" {
		namespace = c.namespace
	}

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

	go func() {
		defer close(logChan)
		defer stream.Close()

		reader := bufio.NewReader(stream)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := reader.ReadString('\n')
				if err != nil {
					if err != io.EOF {
						logChan <- LogLine{Error: err}
					}
					return
				}
				logChan <- LogLine{Content: line}
			}
		}
	}()

	return logChan, nil
}

func (c *Client) GetPodLogSnapshot(ctx context.Context, namespace, podName string, opts LogOptions) (string, error) {
	if namespace == "" {
		namespace = c.namespace
	}

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
