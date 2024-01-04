package ssh

import "io"

type Command struct {
	Name  string
	Idx   int
	Total int
}

type Context[T any] struct {
	writer       io.Writer
	reader       io.Reader
	closeConn    func() error
	closeSession func() error

	CmdChan chan T
	OutChan chan string
	ErrChan chan error

	ParseCmd func(cmd string, idx, total int) T
}

func Init[T any](username, address, privateKey string) (Context[T], error) {
	var err error
	ctx := Context[T]{
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

	ctx.CmdChan = make(chan T)
	ctx.OutChan = make(chan string)
	ctx.ErrChan = make(chan error)

	return ctx, nil
}

func (ctx Context[T]) Close() error {
	if err := ctx.closeSession(); err != nil {
		return err
	}
	if err := ctx.closeConn(); err != nil {
		return err
	}
	return nil
}
