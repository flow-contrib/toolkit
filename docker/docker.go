package docker

import (
	"bytes"
	goctx "context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

type Command struct {
	Environment []string `json:"environment"`
	Command     []string `json:"command"`
	Stdin       string   `json:"stdin"`
}

type OutputValue struct {
	Host    string  `json:"host"`
	Command Command `json:"command"`
	Output  string  `json:"output"`
}

func init() {
	flow.RegisterHandler("toolkit.docker.container.exec", Exec)
	flow.RegisterHandler("toolkit.docker.container.log", Log)
}

func Exec(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	host := conf.GetString("host", "tcp://192.168.99.100:2376")
	tlsVerify := conf.GetString("tls-verify", "1")
	certPath := conf.GetString("cert-path")

	container := conf.GetString("container")

	if len(container) == 0 {
		err = fmt.Errorf("please input container name or id")
		return
	}

	command := conf.GetStringList("command")
	stdin := conf.GetString("stdin")
	envs := conf.GetStringList("environment")
	quiet := conf.GetBoolean("quiet")
	timeout := conf.GetTimeDuration("timeout")

	if len(command) == 0 {
		fmt.Errorf("config of command could not be empty, e.g.: command = [\"/bin/bash\"]")
		return
	}

	errWriter := bytes.NewBuffer(nil)
	outWriter := bytes.NewBuffer(nil)

	var stdErr, stdOut io.Writer

	if quiet {
		stdErr = errWriter
		stdOut = outWriter
	} else {
		stdErr = io.MultiWriter(errWriter, os.Stderr)
		stdOut = io.MultiWriter(outWriter, os.Stdout)
	}

	err = os.Setenv("DOCKER_TLS_VERIFY", tlsVerify)
	if err != nil {
		return
	}

	err = os.Setenv("DOCKER_HOST", host)
	if err != nil {
		return
	}

	err = os.Setenv("DOCKER_CERT_PATH", certPath)
	if err != nil {
		return
	}

	docker, err := NewEnvDocker()
	if err != nil {
		return
	}

	cmd := Command{
		Environment: envs,
		Command:     command,
		Stdin:       stdin,
	}

	executor := docker.Attach(container, cmd)

	executor.Stderr = stdErr
	executor.Stdout = stdOut
	executor.WithStdIn([]byte(stdin))

	c := goctx.Background()

	if timeout > 0 {
		var cancel goctx.CancelFunc
		c, cancel = goctx.WithTimeout(goctx.Background(), timeout)
		defer cancel()
	}

	executor.Ctx = c

	err = executor.Exec()
	if err != nil {
		if errWriter.Len() > 0 {
			err = fmt.Errorf("execute command on docker %s error: %s, details: %s", host, err.Error(), strings.TrimSuffix(errWriter.String(), "\n"))
			return
		}
		err = fmt.Errorf("execute command on docker %s error: %s", host, err.Error())
		return
	}

	outputName := conf.GetString("output.name")

	outputData, err := json.Marshal(OutputValue{
		Host:    host,
		Command: cmd,
		Output:  strings.TrimSuffix(outWriter.String(), "\n"),
	})

	if err != nil {
		return
	}

	flow.AppendOutput(ctx, flow.NameValue{
		Name:  outputName,
		Value: outputData,
		Tags:  []string{"toolkit", "docker", "exec"},
	})

	return
}

func Log(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	host := conf.GetString("host", "tcp://192.168.99.100:2376")
	tlsVerify := conf.GetString("tls-verify", "1")
	certPath := conf.GetString("cert-path")

	container := conf.GetString("container")

	if len(container) == 0 {
		err = fmt.Errorf("please input container name or id")
		return
	}

	quiet := conf.GetBoolean("quiet")
	timeout := conf.GetTimeDuration("timeout")

	outWriter := bytes.NewBuffer(nil)

	var stdOut io.Writer

	if quiet {
		stdOut = outWriter
	} else {
		stdOut = io.MultiWriter(outWriter, os.Stdout)
	}

	err = os.Setenv("DOCKER_TLS_VERIFY", tlsVerify)
	if err != nil {
		return
	}

	err = os.Setenv("DOCKER_HOST", host)
	if err != nil {
		return
	}

	err = os.Setenv("DOCKER_CERT_PATH", certPath)
	if err != nil {
		return
	}

	cli, err := client.NewEnvClient()
	if err != nil {
		return
	}

	c := goctx.Background()

	if timeout > 0 {
		var cancel goctx.CancelFunc
		c, cancel = goctx.WithTimeout(goctx.Background(), timeout)
		defer cancel()
	}

	options := types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}

	out, err := cli.ContainerLogs(c, container, options)

	if err != nil {
		return
	}

	_, err = io.Copy(stdOut, out)

	if err != nil {
		return
	}

	outputName := conf.GetString("output.name")

	if err != nil {
		return
	}

	outputData, err := json.Marshal(map[string]string{"log": outWriter.String()})

	if err != nil {
		return
	}

	flow.AppendOutput(ctx, flow.NameValue{
		Name:  outputName,
		Value: outputData,
		Tags:  []string{"toolkit", "docker", "log"},
	})

	return
}
