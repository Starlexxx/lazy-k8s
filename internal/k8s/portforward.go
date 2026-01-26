package k8s

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ErrPortForwarderNotStarted is returned when trying to get ports before starting the forwarder.
var ErrPortForwarderNotStarted = errors.New("port forwarder not started")

type PortForwardOptions struct {
	Namespace  string
	PodName    string
	LocalPort  int
	RemotePort int
	StopCh     chan struct{}
	ReadyCh    chan struct{}
}

type PortForwarder struct {
	client  *Client
	options PortForwardOptions
	pf      *portforward.PortForwarder
}

func (c *Client) NewPortForwarder(opts PortForwardOptions) (*PortForwarder, error) {
	if opts.Namespace == "" {
		opts.Namespace = c.namespace
	}

	return &PortForwarder{
		client:  c,
		options: opts,
	}, nil
}

func (pf *PortForwarder) Start() error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		pf.options.Namespace, pf.options.PodName)

	hostIP := pf.client.restConfig.Host

	serverURL, err := url.Parse(hostIP)
	if err != nil {
		return err
	}

	serverURL.Path = path

	transport, upgrader, err := spdy.RoundTripperFor(pf.client.restConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: transport},
		http.MethodPost,
		serverURL,
	)

	ports := []string{fmt.Sprintf("%d:%d", pf.options.LocalPort, pf.options.RemotePort)}

	fw, err := portforward.New(dialer, ports, pf.options.StopCh, pf.options.ReadyCh, nil, nil)
	if err != nil {
		return err
	}

	pf.pf = fw

	return fw.ForwardPorts()
}

func (pf *PortForwarder) GetPorts() ([]portforward.ForwardedPort, error) {
	if pf.pf == nil {
		return nil, ErrPortForwarderNotStarted
	}

	return pf.pf.GetPorts()
}

func (pf *PortForwarder) Stop() {
	if pf.options.StopCh != nil {
		close(pf.options.StopCh)
	}
}

func (pf *PortForwarder) LocalPort() int {
	return pf.options.LocalPort
}

func (pf *PortForwarder) RemotePort() int {
	return pf.options.RemotePort
}

func (pf *PortForwarder) PodName() string {
	return pf.options.PodName
}
