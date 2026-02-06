package k8s

import (
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type ExecOptions struct {
	Namespace string
	PodName   string
	Container string
	Command   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	TTY       bool
}

func (c *Client) Exec(opts ExecOptions) error {
	if opts.Namespace == "" {
		opts.Namespace = c.namespace
	}

	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(c.Context(), remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})
}

func (c *Client) GetContainerShell(namespace, podName, container string) []string {
	shells := [][]string{
		{"/bin/bash"},
		{"/bin/sh"},
		{"sh"},
	}

	return shells[0]
}
