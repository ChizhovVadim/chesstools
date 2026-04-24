package uci

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// owner
type Process struct {
	cmd *exec.Cmd
	r   io.ReadCloser
	w   io.WriteCloser
}

func Start(path string, arg string) (*Process, error) {
	var args []string
	if arg != "" {
		args = strings.Fields(arg)
	}
	//TODO use exec.CommandContext()?
	var cmd = exec.Command(path, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return &Process{
		r:   out,
		w:   in,
		cmd: cmd,
	}, nil
}

func (p *Process) Close() error {
	fmt.Fprintln(p.w, "quit")
	return p.cmd.Wait()
}

func (p *Process) Reader() io.Reader {
	return p.r
}

func (p *Process) Writer() io.Writer {
	return p.w
}
