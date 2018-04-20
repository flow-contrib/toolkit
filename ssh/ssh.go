package ssh

import (
	"fmt"
	"os"
	// "fmt"
	goctx "context"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

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
	envs := conf.GetStringList("envs")
	stdin := conf.GetString("stdin")

	if len(command) == 0 {
		fmt.Errorf("config of command could not be empty, e.g.: command = [\"/bin/bash\"]")
		return
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

		Stderr: os.Stdin,
		Stdout: os.Stdout,
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

	return
}
