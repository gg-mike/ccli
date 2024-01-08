package docker

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gg-mike/ccli/pkg/runner"
)

type Manager struct {
	clients map[string]*Client
}

var manager *Manager

func Get() *Manager {
	if manager == nil {
		panic("manager is not initialized")
	}
	return manager
}

func Init() error {
	if manager != nil {
		panic("manager is already initialized")
	}

	manager = &Manager{
		clients: map[string]*Client{},
	}

	return nil
}

func NewClient(host string) error {
	conn, err := newClient(host)
	if err != nil {
		return err
	}
	Get().clients[host] = conn
	return nil
}

func DeleteClient(host string) error {
	if err := Get().clients[host].client.Close(); err != nil {
		return err
	}
	delete(Get().clients, host)
	return nil
}

func NewRunner(host, imageName string) (*runner.Runner, error) {
	conn, ok := Get().clients[host]
	if !ok {
		var err error
		conn, err = newClient(host)
		if err != nil {
			return &runner.Runner{}, err
		}
		Get().clients[host] = conn
	}
	_runner, err := newRunner(conn.client, imageName)
	if err != nil {
		return &runner.Runner{}, err
	}
	return _runner, nil
}

func Shutdown() error {
	if len(Get().clients) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(Get().clients))

	for _, cli := range Get().clients {
		wg.Add(1)
		go cli.Shutdown(errChan, &wg)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	errMsg := []string{}
	for result := range errChan {
		errMsg = append(errMsg, result.Error())
	}

	if len(errMsg) != 0 {
		return fmt.Errorf("manager close: %d clients closed with error\n%s", len(errMsg), strings.Join(errMsg, "\n"))
	}
	return nil
}
