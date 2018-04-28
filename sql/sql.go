package pwgen

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"text/template"

	"github.com/elgs/gosqljson"
	"github.com/gogap/config"
	"github.com/gogap/context"
	"github.com/gogap/flow"
)

var (
	Tags = []string{"toolkit", "sql"}
)

type sqlConfig struct {
	Driver   string
	User     string
	Password string
	Host     string
	Port     int
	Db       string
	Charset  string
	Location string
}

func (p *sqlConfig) DSN() string {
	if p.Driver == "mysql" {
		loc := strings.Replace(p.Location, "/", "%2F", 1)
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&loc=%s", p.User, p.Password, p.Host, p.Port, p.Db, p.Charset, loc)
	} else if p.Driver == "postgres" {
		if len(p.Db) == 0 {
			p.Db = "postgres"
		}
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Db)
	}

	panic("unsupport driver:" + p.Driver)
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
		Driver:   conf.GetString("driver", "mysql"),
		Host:     conf.GetString("host", "localhost"),
		Port:     int(conf.GetInt32("port", 3306)),
		Db:       conf.GetString("db"),
		User:     conf.GetString("user", "root"),
		Password: conf.GetString("password"),
		Charset:  conf.GetString("charset", "utf8"),
		Location: conf.GetString("loc"),
	}

	sqlQuery := conf.GetString("sql")

	sqlQuery, err = renderSQL(sqlQuery, conf.GetConfig("variables"))
	if err != nil {
		return
	}

	db, err := sql.Open(sqlConf.Driver, sqlConf.DSN())

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
		Driver:   conf.GetString("driver", "mysql"),
		Host:     conf.GetString("host", "localhost"),
		Port:     int(conf.GetInt32("port", 3306)),
		Db:       conf.GetString("db"),
		User:     conf.GetString("user", "root"),
		Password: conf.GetString("password"),
		Charset:  conf.GetString("charset", "utf8"),
		Location: conf.GetString("loc"),
	}

	sqlExec := conf.GetString("sql")

	sqlExec, err = renderSQL(sqlExec, conf.GetConfig("variables"))
	if err != nil {
		return
	}

	db, err := sql.Open(sqlConf.Driver, sqlConf.DSN())

	if err != nil {
		return
	}

	defer db.Close()

	isTrans := conf.GetBoolean("tx", true)

	sqls := strings.Split(sqlExec, ";")

	if isTrans {

		var tx *sql.Tx
		tx, err = db.Begin()
		if err != nil {
			return
		}

		for i := 0; i < len(sqls); i++ {
			sql := trimSQL(sqls[i])
			if len(sql) == 0 {
				continue
			}
			_, err = gosqljson.ExecTx(tx, sql)
			if err != nil {
				tx.Rollback()
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			return
		}
	} else {
		for i := 0; i < len(sqls); i++ {
			sql := trimSQL(sqls[i])
			if len(sql) == 0 {
				continue
			}
			_, err = gosqljson.ExecDb(db, sql)
			if err != nil {
				return
			}
		}
	}

	return
}

func renderSQL(sqlExec string, varsConf config.Configuration) (string, error) {

	vars := make(map[string]string)

	if !varsConf.IsEmpty() {
		for _, k := range varsConf.Keys() {
			vars[k] = varsConf.GetString(k)
		}
	}

	tmp, err := template.New("sql").Parse(sqlExec)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	err = tmp.Execute(buf, vars)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func trimSQL(sql string) string {
	sql = strings.TrimSuffix(sql, "\n")
	sql = strings.TrimSuffix(sql, ";")

	sql = strings.TrimSpace(sql)

	if len(sql) == 0 {
		return ""
	}

	sql = sql + ";"

	return sql
}
