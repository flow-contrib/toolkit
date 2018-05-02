package net

import (
	"fmt"
	"net"
	"time"

	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
	"github.com/sirupsen/logrus"
)

func init() {
	flow.RegisterHandler("toolkit.net.tcp.address.check", TCPAddressCheck)
}

func TCPAddressCheck(ctx context.Context, conf config.Configuration) (err error) {
	addr := conf.GetString("address")

	if len(addr) == 0 {
		err = fmt.Errorf("check addr is empty")
		return
	}

	interval := conf.GetTimeDuration("interval", time.Second*3)
	times := int(conf.GetInt32("times", 10))

	if times == 0 {

		for {
			conn, e := net.Dial("tcp", addr)
			if e != nil {
				logrus.WithField("ADDRESS", addr).WithField("CONTENT", e.Error()).Infoln("Tcp address check failure")
				time.Sleep(interval)
				continue
			}
			logrus.WithField("ADDRESS", addr).Infoln("Tcp address check success")
			conn.Close()
			return
		}

		return
	}

	for i := 0; i < times; i++ {

		conn, e := net.Dial("tcp", addr)
		if e != nil {
			logrus.WithField("ADDRESS", addr).WithField("CONTENT", e.Error()).Infoln("Tcp address check failure")
			time.Sleep(interval)
			continue
		}
		logrus.WithField("ADDRESS", addr).Infoln("Tcp address check success")
		conn.Close()
		return
	}

	err = fmt.Errorf("check address '%s' timeout", addr)

	return
}
