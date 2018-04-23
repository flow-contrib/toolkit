package pwgen

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chr4/pwgen"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

var (
	Tags = []string{"toolkit", "pwgen"}
)

type Password struct {
	Name        string   `json:"name"`
	Length      int      `json:"length"`
	Encoding    string   `json:"encoding"`
	Plain       string   `json:"plain"`
	Encoded     string   `json:"encoded"`
	HasSymbols  bool     `json:"symbols"`
	Environment []string `json:"environment"`
	envPrefix   string
	exportToEnv bool
	output      bool
}

func (p *Password) Generate() (err error) {

	if p.Length <= 0 {
		p.Length = 16
	}

	pwd := ""

	if p.HasSymbols {
		pwd = pwgen.AlphaNumSymbols(p.Length)
	} else {
		pwd = pwgen.AlphaNum(p.Length)
	}

	switch p.Encoding {

	case "sha256":
		{
			var hashed [32]byte
			hashed = sha256.Sum256([]byte(pwd))

			p.Plain = pwd
			p.Encoded = fmt.Sprintf("%0x", hashed)
		}
	case "sha512":
		{
			var hashed [64]byte
			hashed = sha512.Sum512([]byte(pwd))

			p.Plain = pwd
			p.Encoded = fmt.Sprintf("%0x", hashed)
		}
	case "md5":
		{
			var hashed [16]byte
			hashed = md5.Sum([]byte(pwd))

			p.Plain = pwd
			p.Encoded = fmt.Sprintf("%0x", hashed)
		}
	case "base64":
		{
			p.Plain = pwd
			p.Encoded = base64.StdEncoding.EncodeToString([]byte(pwd))
		}
	default:
		p.Encoding = "plain"
		p.Plain = pwd
		p.Encoded = pwd
	}

	return
}

func (p *Password) ExportToEnv() {
	if p.exportToEnv {
		os.Setenv(p.envPrefix+"_PLAIN", p.Plain)
		os.Setenv(p.envPrefix+"_ENCODED", p.Encoded)
	}
}

func (p Password) AppendOutput(ctx context.Context) (err error) {
	if !p.output {
		return
	}

	data, err := json.Marshal(p)
	if err != nil {
		return
	}

	flow.AppendOutput(ctx, flow.NameValue{Name: p.Name, Value: data, Tags: Tags})

	return
}

func init() {
	flow.RegisterHandler("toolkit.pwgen.generate", Generate)
}

func Generate(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	names := conf.Keys()

	if len(names) == 0 {
		return
	}

	var pwds []Password

	for _, name := range names {

		pwdConf := conf.GetConfig(name)

		if pwdConf.IsEmpty() {
			continue
		}

		outputName := pwdConf.GetString("name", name)

		originalOutputs := flow.FindOutput(ctx, outputName, Tags...)

		if len(originalOutputs) == 0 {

			pwdlen := pwdConf.GetInt32("len", 16)
			encoding := pwdConf.GetString("encoding", "")
			symbols := pwdConf.GetBoolean("symbols", false)
			env := pwdConf.GetBoolean("env")

			pwd := Password{
				Name:        outputName,
				Length:      int(pwdlen),
				Encoding:    encoding,
				HasSymbols:  symbols,
				exportToEnv: env,
				output:      true,
			}

			if env {
				envPrefix := toEnvFomart(outputName)
				pwd.envPrefix = envPrefix
				pwd.Environment = []string{envPrefix + "_PLAIN", envPrefix + "_ENCODED"}
			}

			err = pwd.Generate()
			if err != nil {
				return
			}

			pwds = append(pwds, pwd)

		} else if len(originalOutputs) > 1 {
			err = fmt.Errorf("conflict of output name: %s, tags: %s", outputName, strings.Join(Tags, ","))
			return
		} else {

			var pwd Password
			errUnmarshal := json.Unmarshal(originalOutputs[0].Value, &pwd)
			if errUnmarshal != nil {
				err = fmt.Errorf("convert output name: %s, tags: %s to password failure: %s", outputName, strings.Join(Tags, ","), errUnmarshal)
				return
			}

			if len(pwd.Name) == 0 {
				err = fmt.Errorf("convert output name: %s, tags: %s to password failure: it was empty content", outputName, strings.Join(Tags, ","))
				return
			}

			pwd.output = false
			pwds = append(pwds, pwd)
		}
	}

	for i := 0; i < len(pwds); i++ {
		err = pwds[i].AppendOutput(ctx)
		if err != nil {
			return
		}

		pwds[i].ExportToEnv()
	}

	return
}

func toEnvFomart(key string) string {
	key = strings.ToUpper(key)
	key = strings.Replace(key, "-", "_", -1)
	return key
}
