package ssh

import (
	"bytes"
	goctx "context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/pkg/sftp"
)

type OutputValue struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`

	Command Command `json:"command"`

	Output string `json:"output"`
}

func init() {
	flow.RegisterHandler("toolkit.ssh.command.run", Run)
	flow.RegisterHandler("toolkit.ssh.file.upload", Upload)
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

func Upload(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	files := conf.GetStringList("files")

	if len(files) == 0 {
		return
	}

	user := conf.GetString("user")
	password := conf.GetString("password")
	host := conf.GetString("host", "localhost")
	port := conf.GetString("port", "22")
	identityFile := conf.GetString("identity-file")
	connectRetries := conf.GetInt32("connect-retries", 3)
	maxPacket := conf.GetInt32("max-packet", 20480)
	quiet := conf.GetBoolean("quiet")
	ignore := conf.GetStringList("ignore")

	cli := Client{
		Config: Config{
			User:           user,
			Password:       password,
			Host:           host,
			Port:           port,
			IdentityFile:   identityFile,
			ConnectRetries: int(connectRetries),
		},
	}

	err = cli.Connect()

	if err != nil {
		return
	}

	defer cli.Cleanup()

	var sftpClient *sftp.Client
	sftpClient, err = sftp.NewClient(cli.client, sftp.MaxPacket(int(maxPacket)))

	mapFiles := map[string]string{}
	fileOrder := []string{}

	for _, file := range files {
		items := strings.Split(file, ":")

		if len(items) != 2 {
			err = fmt.Errorf("file format error, should be 'localfile:remotefile'")
			return
		}

		mapFiles[items[0]] = items[1]
		fileOrder = append(fileOrder, items[0])
	}

	for _, file := range fileOrder {

		var fi os.FileInfo
		fi, err = os.Stat(file)
		if err != nil {
			return
		}

		if fi.IsDir() {

			remoteDirRoot := mapFiles[file]
			localDirBase := filepath.Base(file)

			err = filepath.Walk(file,
				func(path string, info os.FileInfo, walkErr error) error {

					relPath, e := filepath.Rel(file, path)
					if e != nil {
						return e
					}

					remotePath := filepath.Join(remoteDirRoot, localDirBase, relPath)

					for _, pattern := range ignore {
						matched, matchErr := filepath.Match(pattern, info.Name())
						if matched {
							if info.IsDir() {
								return filepath.SkipDir
							}
							return nil
						}

						if matchErr != nil {
							return matchErr
						}
					}

					if info.IsDir() {
						e = cli.Exec(fmt.Sprintf("mkdir -p '%s'", remotePath))
						if e != nil {
							return e
						}
						return nil
					}

					e = uploadFile(sftpClient, quiet, maxPacket, info, path, remotePath)
					if e != nil {
						return e
					}

					return nil
				},
			)
			if err != nil {
				return
			}

			continue
		}

		err = uploadFile(sftpClient, quiet, maxPacket, fi, file, mapFiles[file])
		if err != nil {
			return
		}
	}

	return
}

func uploadFile(sftpClient *sftp.Client, quiet bool, maxPacket int32, fi os.FileInfo, localFilename, remoateFilename string) (err error) {

	totalSize := fi.Size()

	var localFile *os.File
	localFile, err = os.Open(localFilename)
	if err != nil {
		return
	}
	defer localFile.Close()

	var remoteFile *sftp.File
	remoteFile, err = sftpClient.Create(remoateFilename)

	if err != nil {
		err = fmt.Errorf("create remote file failure, file: %s, error: %s", remoateFilename, err)
		return
	}

	defer remoteFile.Close()

	buf := make([]byte, maxPacket)
	readed := 0

	for {
		n, eRead := localFile.Read(buf)
		readed += n

		if !quiet {
			fmt.Printf("%0.1f%% (%s -> %s)\r", (float64(readed)/float64(totalSize))*100, localFilename, remoateFilename)
		}

		if eRead != nil {
			if eRead == io.EOF {
				break
			} else {
				err = fmt.Errorf("read buf from local file failure, file: %s, error: %s", localFilename, eRead)
				return
			}
		}

		_, eWrite := remoteFile.Write(buf)
		if eWrite != nil {
			err = fmt.Errorf("write buf to remote file failure, file: %s, error: %s", remoateFilename, eWrite)
			return
		}
	}

	if !quiet {
		fmt.Printf("\n\r")
	}

	err = remoteFile.Chmod(fi.Mode())

	if err != nil {
		return
	}

	return
}
