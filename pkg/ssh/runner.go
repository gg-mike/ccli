package ssh

import "github.com/gg-mike/ccli/pkg/runner"

func NewRunner(username, address, privateKey string) (*runner.Runner, error) {
	var err error

	conn, err := NewConnection(username, address, privateKey)
	if err != nil {
		return &runner.Runner{}, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return &runner.Runner{}, err
	}

	w, err := session.StdinPipe()
	if err != nil {
		return &runner.Runner{}, err
	}
	r, err := session.StdoutPipe()
	if err != nil {
		return &runner.Runner{}, err
	}

	if err = session.Shell(); err != nil {
		return &runner.Runner{}, err
	}

	_runner := runner.NewRunner(w, r)
	_runner.OnShutdown = func() error {
		if err := session.Close(); err != nil {
			return err
		}
		return conn.Close()
	}

	return _runner, nil
}
