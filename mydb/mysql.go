package mydb

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tkstorm/audit_engine/tool"
	"log"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Pass     string
	Protocol string
	DbName   string
}

type DbMysql struct {
	Dbcf Config
	Db   sql.DB
}

func (mdb *DbMysql) init() {
	log.Println("mysql init...")
}

func (mdb *DbMysql) Connect(dbcf Config) {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", dbcf.User, dbcf.Pass, dbcf.Protocol, dbcf.Host, dbcf.Port, dbcf.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		tool.FatalLog(err, "connect to mysql fail")
	}
	mdb.Db = *db
}

func (mdb *DbMysql) Close() {
	mdb.Db.Close()
}

//连接
func Concat(ids []interface{}) string {
	return strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ",")
}
