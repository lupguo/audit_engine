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

var DB *sql.DB

func init() {
	log.Println("mysql init...")
}

func Connect(dbcf Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", dbcf.User, dbcf.Pass, dbcf.Protocol, dbcf.Host, dbcf.Port, dbcf.DbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		tool.FatalLog(err, "connect to mysql fail")
	}
	return db
}

func Close(db sql.DB) {
	db.Close()
}

//连接
func Concat(ids []interface{}) string {
	return strings.Join(strings.Split(strings.Repeat("?", len(ids)), ""), ",")
}
