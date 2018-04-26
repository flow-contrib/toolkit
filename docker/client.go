package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/flow-contrib/toolkit/utils/shell"
	"github.com/sirupsen/logrus"
	"io"
	"time"
)

type Docker struct {
	client *client.Client
}

func NewDocker(cli *client.Client) (*Docker, error) {
	return &Docker{
		client: cli,
	}, nil
}

func NewEnvDocker() (*Docker, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	return &Docker{client: cli}, nil
}

func (p *Docker) Attach(container string, cmd Command) *ContainerExecutor {
	return &ContainerExecutor{
		client:    p.client,
		Ctx:       context.TODO(),
		Stdout:    bytes.NewBuffer(nil),
		Stderr:    bytes.NewBuffer(nil),
		container: container,
		command:   cmd,
	}
}

type ContainerExecutor struct {
	client    *client.Client
	Ctx       context.Context
	Stdout    io.Writer
	Stderr    io.Writer
	container string
	command   Command
	intput    io.Reader
}

func (p *ContainerExecutor) WithStdIn(script []byte) *ContainerExecutor {
	p.intput = bytes.NewReader(script)
	return p
}

func (p *ContainerExecutor) Exec() (err error) {

	options := types.ExecConfig{
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          p.command.Command,
	}

	exec, err := p.client.ContainerExecCreate(p.Ctx, p.container, options)
	if err != nil {
		return
	}

	hijacked, err := p.client.ContainerExecAttach(p.Ctx, exec.ID, options)
	if err != nil {
		return
	}
	defer hijacked.Close()

	attachCh := make(chan error, 2)

	go func() {
		_, err := stdcopy.StdCopy(p.Stdout, p.Stderr, hijacked.Reader)
		if err != nil {
			attachCh <- err
		}
	}()

	go func() {

		var envVariables bytes.Buffer
		for _, keyValue := range p.command.Environment {
			envVariables.WriteString("export " + shell.Escape(keyValue) + "\n")
		}

		stdIn := io.MultiReader(
			&envVariables,
			bytes.NewBufferString(p.command.Stdin),
		)

		_, err := io.Copy(hijacked.Conn, stdIn)
		hijacked.CloseWrite()
		if err != nil {
			attachCh <- err
		}
	}()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- p.waitForExec(p.Ctx, exec.ID)
	}()

	select {
	case <-p.Ctx.Done():
		err = errors.New("Aborted")
	case err = <-attachCh:
		logrus.Debugln("Exec", exec.ID, "finished with", err)
	case err = <-waitCh:
		logrus.Debugln("Exec", exec.ID, "finished with", err)
	}

	// if err != nil {
	// 	err = fmt.Errorf("%s\n%s", p.Stderr.String(), err)
	// 	return
	// }

	return

}

func (p *ContainerExecutor) waitForExec(ctx context.Context, id string) error {
	logrus.Debugln("Waiting for exec", id, "...")

	retries := 0

	for {
		exec, err := p.client.ContainerExecInspect(ctx, id)

		if err != nil {
			if client.IsErrNotFound(err) {
				return err
			}

			if retries > 3 {
				return err
			}

			retries++
			time.Sleep(time.Second)
			continue
		}

		retries = 0

		if exec.Running {
			time.Sleep(time.Second)
			continue
		}

		if exec.ExitCode != 0 {
			return fmt.Errorf("exit code %d", exec.ExitCode)
		}

		return nil
	}
}
