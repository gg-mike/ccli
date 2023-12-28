package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (ctx *Context) Run(commands []string) {
	running := true

	go ctx.run(commands)

	for running {
		select {
		case cmd := <-ctx.cmdChan:
			ctx.OnCmd(cmd)
		case out := <-ctx.outChan:
			ctx.OnOut(out)
		case err := <-ctx.errChan:
			ctx.OnErr(err)
			running = false
		}
	}
}

func (ctx *Context) run(commands []string) {
	OUT_CMD_TERM := "Ua&&Bi9G*TjbPF62oGa4"
	ERR_CMD_TERM := "!N3o#F4SPZ&UDxybohUT"

	total := len(commands)
	in := make(chan string)
	inStatus := make(chan error)

	outReader := SyncReader{
		source:     ctx.reader,
		term:       OUT_CMD_TERM,
		errTerm:    ERR_CMD_TERM,
		ready:      make(chan any),
		scanStatus: make(chan error),
	}

	defer close(in)
	defer close(outReader.ready)

	go func() {
		for command := range in {
			cmd := []byte(fmt.Sprintf("%s 2>&1 && echo '%s' || echo '%s'\n", command, OUT_CMD_TERM, ERR_CMD_TERM))
			_, err := ctx.writer.Write(cmd)
			inStatus <- err
		}
	}()

	go scan(outReader, ctx.outChan)

	// Read welcome message
	if err := runCommand("echo", in, inStatus, outReader); err != nil {
		// Error during connection like 'mesg: ttyname failed: Inappropriate ioctl for device'
		if err.Error() != "command ended with error" {
			ctx.errChan <- err
		}
	}

	for i, command := range commands {
		ctx.cmdChan <- ctx.ParseCmd(command, i, total)
		if err := runCommand(command, in, inStatus, outReader); err != nil {
			ctx.errChan <- err
		}
	}
	ctx.errChan <- nil
}

type SyncReader struct {
	source     io.Reader
	term       string
	errTerm    string
	ready      chan any
	scanStatus chan error
}

func scan(reader SyncReader, outChan chan string) {
	scanner := bufio.NewScanner(reader.source)
	for range reader.ready {
		for {
			if tkn := scanner.Scan(); tkn {
				text := scanner.Text()
				if strings.Contains(text, reader.term) {
					break
				} else if strings.Contains(text, reader.errTerm) {
					reader.scanStatus <- errors.New("command ended with error")
					return
				} else {
					outChan <- text
				}
			} else {
				reader.scanStatus <- scanner.Err()
				return
			}
		}
		reader.scanStatus <- nil
	}
}

func runCommand(command string, in chan string, inStatus chan error, reader SyncReader) error {
	in <- command
	if err := <-inStatus; err != nil {
		return err
	}
	reader.ready <- true
	if err := <-reader.scanStatus; err != nil {
		return err
	}
	return nil
}
