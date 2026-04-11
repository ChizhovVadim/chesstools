package uci

import (
	"fmt"
	"os/exec"
)

type Process struct {
	name    string
	path    string
	args    []string
	options []Option
	cmd     *exec.Cmd
	*Service
}

func NewProcess(
	name string,
	path string,
	args []string,
	options []Option,
) *Process {
	return &Process{
		name:    name,
		path:    path,
		args:    args,
		options: options,
	}
}

func (p *Process) Name() string {
	return p.name
}

func (p *Process) Init() error {
	var cmd = exec.Command(p.path, p.args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	p.cmd = cmd
	p.Service = NewService(out, in)
	//p.Service = NewService(out, io.MultiWriter(in, os.Stdout)) //Для отладки
	p.Uci()
	for _, option := range p.options {
		p.SetOption(option)
	}
	if !p.IsReady() {
		return fmt.Errorf("engine not ready")
	}
	return nil
}

func (p *Process) Close() error {
	if p.cmd == nil {
		return nil
	}
	p.Quit()
	return p.cmd.Wait()
}
