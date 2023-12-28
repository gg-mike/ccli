package ssh

import "io"

type Command struct {
	Name  string
	Idx   int
	Total int
}

type Context struct {
	writer       io.Writer
	reader       io.Reader
	cmdChan      chan any
	outChan      chan string
	errChan      chan error
	closeConn    func() error
	closeSession func() error

	ParseCmd func(cmd string, idx, total int) any

	OnCmd func(any)
	OnOut func(string)
	OnErr func(error)
}

func Init(username, address, privateKey string) (Context, error) {
	var err error
	ctx := Context{
		closeConn:    func() error { return nil },
		closeSession: func() error { return nil },
	}
	conn, err := NewConnection(username, address, privateKey)
	if err != nil {
		return ctx, err
	}
	ctx.closeConn = conn.Close

	session, err := conn.NewSession()
	if err != nil {
		return ctx, err
	}
	ctx.closeSession = session.Close

	ctx.writer, err = session.StdinPipe()
	if err != nil {
		return ctx, err
	}
	ctx.reader, err = session.StdoutPipe()
	if err != nil {
		return ctx, err
	}

	if err = session.Shell(); err != nil {
		return ctx, err
	}

	ctx.cmdChan = make(chan any)
	ctx.outChan = make(chan string)
	ctx.errChan = make(chan error)

	return ctx, nil
}
