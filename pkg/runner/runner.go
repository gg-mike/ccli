package runner

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var ErrBuildFailed = errors.New("build failed")

type Runner struct {
	writer  *bufio.Writer
	scanner *bufio.Scanner

	OnCmd      func(cmd string, idx int, total int)
	OnOut      func(out string)
	OnShutdown func() error
}

func NewRunner(writer io.Writer, reader io.Reader) *Runner {
	return &Runner{
		writer:  bufio.NewWriter(writer),
		scanner: bufio.NewScanner(reader),
	}
}

func (r *Runner) Run(commands []string) error {
	OUT_CMD_TERM := "Ua&&Bi9G*TjbPF62oGa4"
	ERR_CMD_TERM := "!N3o#F4SPZ&UDxybohUT"

	total := len(commands)

	for idx, command := range commands {
		r.OnCmd(command, idx, total)

		cmd := fmt.Sprintf("%s 2>&1 && echo '%s' || echo '%s'\n", command, OUT_CMD_TERM, ERR_CMD_TERM)
		_, err := r.writer.WriteString(cmd)
		if err != nil {
			return err
		}

		err = r.writer.Flush()
		if err != nil {
			return err
		}

		for {
			if tkn := r.scanner.Scan(); tkn {
				text := cleanupScan(r.scanner.Bytes())
				if strings.Contains(text, OUT_CMD_TERM) {
					break
				} else if strings.Contains(text, ERR_CMD_TERM) {
					return ErrBuildFailed
				} else {
					r.OnOut(text)
				}
			} else {
				return r.scanner.Err()
			}
		}
	}

	return nil
}

func (r *Runner) Shutdown() error {
	return r.OnShutdown()
}

func cleanupScan(value []byte) string {
	if len(value) == 0 || value[0] != 1 {
		return string(value)
	}
	return string(value[8:])
}
