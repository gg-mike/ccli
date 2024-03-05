package docker

import (
	"sync"

	"github.com/docker/docker/client"
)

type Client struct {
	host   string
	client *client.Client
}

func newClient(host string) (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return &Client{}, err
	}
	return &Client{
		host:   host,
		client: cli,
	}, nil
}

func (c *Client) Shutdown(errChan chan error, wg *sync.WaitGroup) {
	if err := c.client.Close(); err != nil {
		errChan <- err
	}
	wg.Done()
}
