package docker

import (
	"bufio"
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gg-mike/ccli/pkg/runner"
)

func newRunner(cli *client.Client, imageName string, privileged bool) (*runner.Runner, error) {
	out, err := cli.ImagePull(context.Background(), imageName, types.ImagePullOptions{})
	if err != nil {
		return &runner.Runner{}, err
	}
	defer out.Close()
	scanner := bufio.NewScanner(out)
	for {
		if tkn := scanner.Scan(); tkn {
		} else {
			break
		}
	}

	resp, err := cli.ContainerCreate(context.Background(), &container.Config{
		Image:        imageName,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          false,
		AttachStdout: true,
		OpenStdin:    true,
	},
		&container.HostConfig{
			AutoRemove: true,
			Privileged: privileged,
		}, nil, nil, "")
	if err != nil {
		return &runner.Runner{}, err
	}

	conn, err := cli.ContainerAttach(context.Background(), resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
	})
	if err != nil {
		return &runner.Runner{}, err
	}

	err = cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return &runner.Runner{}, err
	}

	_runner := runner.NewRunner(conn.Conn, conn.Conn)
	_runner.OnShutdown = func() error {
		if err := conn.Conn.Close(); err != nil {
			return err
		}
		return cli.ContainerRemove(context.Background(), resp.ID, types.ContainerRemoveOptions{
			Force: true,
		})
	}

	return _runner, nil
}
