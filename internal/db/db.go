package db

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type Config struct {
	User   string
	Passwd string
	Host   string
	Port   int
	DBName string
}

func Open(cfg Config) (*sql.DB, error) {
	mc := mysql.NewConfig()
	mc.User = cfg.User
	mc.Passwd = cfg.Passwd
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mc.DBName = cfg.DBName
	mc.ParseTime = true
	mc.MultiStatements = true

	db, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
