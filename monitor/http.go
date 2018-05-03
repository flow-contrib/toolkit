package monitor

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("toolkit.monitor.http.content.check", HTTPContentCheck)
}

func HTTPContentCheck(ctx context.Context, conf config.Configuration) (err error) {
	url := conf.GetString("url")

	if len(url) == 0 {
		err = errors.New("config of url is empty")
		return
	}

	method := conf.GetString("method", "GET")

	data := conf.GetString("body.data")
	isBase64 := conf.GetBoolean("body.is-base64")
	interval := conf.GetTimeDuration("interval", time.Second*3)
	times := conf.GetTimeDuration("times", 0)

	patterns := conf.GetStringList("patterns")

	body := []byte(data)

	if isBase64 {
		body, err = base64.StdEncoding.DecodeString(data)
		if err != nil {
			err = fmt.Errorf("the check body decode base64 failure: %s", url)
			return
		}
	}

	count := times

	for count > 0 || times == 0 {

		if times > 0 {
			count--
		}

		var ok bool
		ok, err = doHTTPCheck(method, url, body, patterns)

		if err != nil || ok {
			if ok {
				logrus.WithField("url", url).WithField("method", method).Infoln("Checking http content success")
			}
			return
		}

		logrus.WithField("url", url).WithField("method", method).Infoln("Checking http content failure")

		time.Sleep(interval)
	}

	return
}

func doHTTPCheck(method, url string, body []byte, patterns []string) (ok bool, err error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))

	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	respData, err := ioutil.ReadAll(resp.Body)

	for i := 0; i < len(patterns); i++ {
		var matched bool
		matched, err = regexp.MatchString(patterns[i], string(respData))
		if err != nil {
			return
		}

		if !matched {
			return false, nil
		}
	}

	if err != nil {
		return
	}

	ok = true

	return
}
