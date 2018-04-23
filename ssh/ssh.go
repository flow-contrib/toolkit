package ssh

import (
	"bytes"
	goctx "context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

type OutputValue struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`

	Command Command `json:"command"`

	Output string `json:"output"`
}

func init() {
	flow.RegisterHandler("toolkit.ssh.run", Run)
}

func Run(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	user := conf.GetString("user")
	password := conf.GetString("password")
	host := conf.GetString("host", "localhost")
	port := conf.GetString("port", "22")
	identityFile := conf.GetString("identity-file")
	connectRetries := conf.GetInt32("connect-retries", 3)
	timeout := conf.GetTimeDuration("timeout", 0)

	command := conf.GetStringList("command")
	envs := conf.GetStringList("environment")
	stdin := conf.GetString("stdin")

	quiet := conf.GetBoolean("quiet")

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

	cli := Client{
		Config: Config{
			User:           user,
			Password:       password,
			Host:           host,
			Port:           port,
			IdentityFile:   identityFile,
			ConnectRetries: int(connectRetries),
		},

		Stderr: stdErr,
		Stdout: stdOut,
	}

	err = cli.Connect()

	if err != nil {
		return
	}

	defer cli.Cleanup()

	c := goctx.Background()

	if timeout > 0 {
		var cancel goctx.CancelFunc
		c, cancel = goctx.WithTimeout(goctx.Background(), timeout)
		defer cancel()
	}

	cmd := Command{
		Environment: envs,
		Command:     command,
		Stdin:       stdin,
	}

	err = cli.Run(c, cmd)

	if err != nil {
		return
	}

	if errWriter.Len() > 0 {
		err = fmt.Errorf("execute ssh command on server %s@%s:%s error: %s", user, host, port, strings.TrimSuffix(errWriter.String(), "\n"))
		return
	}

	outputName := conf.GetString("output.name")

	outputData, err := json.Marshal(OutputValue{
		Host:    host,
		User:    user,
		Port:    port,
		Command: cmd,
		Output:  strings.TrimSuffix(outWriter.String(), "\n"),
	})

	if err != nil {
		return
	}

	flow.AppendOutput(ctx, flow.NameValue{
		Name:  outputName,
		Value: outputData,
	})

	return
}
