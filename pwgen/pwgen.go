package pwgen

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/chr4/pwgen"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

type Password struct {
	Name       string `json:"name"`
	Length     int    `json:"length"`
	Encoding   string `json:"encoding"`
	Plain      string `json:"plain"`
	Encoded    string `json:"value"`
	HasSymbols bool   `json:"symbols"`
	Env        string `json:"env"`
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
		pwdlen := pwdConf.GetInt32("len", 16)
		encoding := pwdConf.GetString("encoding", "")
		symbols := pwdConf.GetBoolean("symbols", false)
		env := pwdConf.GetString("env")

		pwd := Password{
			Env:        env,
			Name:       outputName,
			Length:     int(pwdlen),
			Encoding:   encoding,
			HasSymbols: symbols,
		}

		err = pwd.Generate()
		if err != nil {
			return
		}

		pwds = append(pwds, pwd)
	}

	for i := 0; i < len(pwds); i++ {
		flow.AppendOutput(ctx, pwds[i].Name, pwds[i])

		if len(pwds[i].Env) > 0 {
			os.Setenv(pwds[i].Env+"_PLAIN", pwds[i].Plain)
			os.Setenv(pwds[i].Env+"_ENCODED", pwds[i].Encoded)
		}
	}

	return
}
