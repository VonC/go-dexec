package dexec

import (
	"errors"
	"io"

	"github.com/fsouza/go-dockerclient"
)

type Docker struct {
	*docker.Client
}

func (d Docker) Command(method Execution, name string, arg ...string) *Cmd {
	return &Cmd{Method: method, Path: name, Args: arg, docker: d}
}

type Cmd struct {
	// Method provides the execution strategy for the context of the Cmd.
	// An instance of Method should not be reused between Cmds.
	Method Execution

	// Path is the path or name of the command in the container.
	Path string

	// Arguments to the command in the container, excluding the command
	// name as the first argument.
	Args []string

	// Env is environment variables to the command. If Env is nil, Run will use
	// Env specified on Method or pre-built container image.
	Env []string

	// Dir specifies the working directory of the command. If Dir is the empty
	// string, Run uses Dir specified on Method or pre-built container image.
	Dir string

	// Stdin specifies the process's standard input.
	// If Stdin is nil, the process reads from the null device (os.DevNull).
	Stdin io.Reader

	// Stdout and Stderr specify the process's standard output and error.
	// If either is nil, they will be redirected to the null device (os.DevNull).
	//
	// TODO concurrency guarantees around calls to Write() if Stdout==Stderr
	Stdout io.Writer
	Stderr io.Writer

	docker  Docker
	started bool
}

func (c *Cmd) Start() error {
	if c.Dir != "" {
		if err := c.Method.setDir(c.Dir); err != nil {
			return err
		}
	}
	if c.Env != nil {
		if err := c.Method.setEnv(c.Env); err != nil {
			return err
		}
	}

	if c.started {
		return errors.New("dexec: already started")
	}
	c.started = true

	if c.Stdin == nil {
		c.Stdin = empty
	}
	if c.Stdout == nil {
		c.Stdout = discard
	}
	if c.Stderr == nil {
		c.Stderr = discard
	}

	cmd := append([]string{c.Path}, c.Args...)
	if err := c.Method.Create(c.docker, cmd); err != nil {
		return err
	}
	if err := c.Method.Run(c.docker, c.Stdin, c.Stdout, c.Stderr); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) Wait() error {
	if !c.started {
		return errors.New("dexec: not started")
	}
	return c.Method.Wait()
	// TODO check container's exit code?
}

func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

func (c *Cmd) StderrPipe() (io.ReadCloser, error) { return nil, nil }
func (c *Cmd) StdinPipe() (io.WriteCloser, error) { return nil, nil }
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) { return nil, nil }
func (c *Cmd) CombinedOutput() ([]byte, error)    { return nil, nil }
func (c *Cmd) Output() ([]byte, error)            { return nil, nil }