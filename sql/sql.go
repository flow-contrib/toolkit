package pwgen

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/elgs/gosqljson"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

var (
	Tags = []string{"toolkit", "sql"}
)

type sqlConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Db       string
	Charset  string
	Location string
}

func (p *sqlConfig) DSN() string {
	loc := strings.Replace(p.Location, "/", "%2F", 1)
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=%s", p.User, p.Password, p.Host, p.Port, p.Db, p.Charset, loc)
}

func init() {
	flow.RegisterHandler("toolkit.sql.query", Query)
	flow.RegisterHandler("toolkit.sql.exec", Exec)
}

func Query(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	sqlConf := sqlConfig{
		Host:     conf.GetString("host", "localhost"),
		Port:     int(conf.GetInt32("port", 3306)),
		Db:       conf.GetString("db"),
		User:     conf.GetString("user", "root"),
		Password: conf.GetString("password"),
		Charset:  conf.GetString("charset", "utf8"),
		Location: conf.GetString("loc"),
	}

	driver := conf.GetString("driver", "mysql")
	sqlQuery := conf.GetString("sql")

	db, err := sql.Open(driver, sqlConf.DSN())

	if err != nil {
		return
	}

	jsonQuery, err := gosqljson.QueryDbToMapJSON(db, "lower", sqlQuery)

	if err != nil {
		return
	}

	outputName := conf.GetString("output.name")

	if len(outputName) > 0 {
		flow.AppendOutput(ctx, flow.NameValue{Name: outputName, Value: []byte(jsonQuery), Tags: Tags})
	}

	return
}

func Exec(ctx context.Context, conf config.Configuration) (err error) {

	if conf.IsEmpty() {
		return
	}

	sqlConf := sqlConfig{
		Host:     conf.GetString("host", "localhost"),
		Port:     int(conf.GetInt32("port", 3306)),
		Db:       conf.GetString("db"),
		User:     conf.GetString("user", "root"),
		Password: conf.GetString("password"),
		Charset:  conf.GetString("charset", "utf8"),
		Location: conf.GetString("loc"),
	}

	driver := conf.GetString("driver", "mysql")
	sqlQuery := conf.GetString("sql")

	db, err := sql.Open(driver, sqlConf.DSN())

	if err != nil {
		return
	}

	tx, err := db.Begin()
	if err != nil {
		return
	}

	_, err = gosqljson.ExecTx(tx, sqlQuery)
	if err != nil {
		tx.Rollback()
		return
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	return
}
