package pwgen

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/howeyc/gopass"
)

var (
	Tags = []string{"toolkit", "readline"}
)

type ReadLineValue struct {
	Name  string `json:"name"`
	Input string `json:"input"`
	Type  string `json:"type"`
}

func init() {
	flow.RegisterHandler("toolkit.readline.text.read", ReadText)
	flow.RegisterHandler("toolkit.readline.password.read", ReadPassword)
}

func ReadText(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	name := conf.GetString("name")
	env := conf.GetBoolean("env")
	prmopt := conf.GetString("prompt", name)
	confirm := conf.GetBoolean("confirm")

reInput:
	fmt.Print(prmopt + ":")
	text := readLine(os.Stdin)

	if confirm {
	reConfirm:
		fmt.Printf("you are input '%s', is it correct? (yes/no):", string(text))
		txtConfirm := readLine(os.Stdin)

		if string(strings.ToUpper(string(txtConfirm))) == "NO" {
			goto reInput
		} else if string(strings.ToUpper(string(txtConfirm))) != "YES" {
			goto reConfirm
		}
	}

	if len(name) > 0 {

		var value []byte
		value, err = json.Marshal(ReadLineValue{Input: string(text), Name: name, Type: "text"})
		if err != nil {
			return
		}

		flow.AppendOutput(ctx, flow.NameValue{Name: name, Value: value, Tags: Tags})

		if env {
			envKey := toEnvFomart(name)
			os.Setenv(envKey, string(text))
		}
	}

	return
}

func ReadPassword(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	name := conf.GetString("name")
	env := conf.GetBoolean("env")
	prompt := conf.GetString("prompt", name)
	confirm := conf.GetBoolean("confirm")

reInput:

	text, err := gopass.GetPasswdPrompt(prompt+":", true, os.Stdin, os.Stdout)
	if err != nil {
		return
	}

	if confirm {

		var text2 []byte
		text2, err = gopass.GetPasswdPrompt(prompt+":", true, os.Stdin, os.Stdout)
		if err != nil {
			return
		}

		if string(text) != string(text2) {
			fmt.Println("!!! twice input did not match")
			goto reInput
		}
	}

	if len(name) > 0 {

		var value []byte
		value, err = json.Marshal(ReadLineValue{Input: string(text), Name: name, Type: "password"})
		if err != nil {
			return
		}

		flow.AppendOutput(ctx, flow.NameValue{Name: name, Value: value, Tags: Tags})

		if env {
			envKey := toEnvFomart(name)
			os.Setenv(envKey, string(text))
		}
	}

	return
}

func readLine(reader io.Reader) []byte {
	buf := bufio.NewReader(reader)
	line, err := buf.ReadBytes('\n')

	for err == nil {
		line = bytes.TrimRight(line, "\n")
		if len(line) > 0 {
			if line[len(line)-1] == 13 { //'\r'
				line = bytes.TrimRight(line, "\r")
			}
			return line
		}
		line, err = buf.ReadBytes('\n')
	}

	if len(line) > 0 {
		return line
	}

	return nil
}

func toEnvFomart(key string) string {
	key = strings.ToUpper(key)
	key = strings.Replace(key, "-", "_", -1)
	return key
}
