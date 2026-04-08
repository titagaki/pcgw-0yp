package db

import (
	"database/sql"
	"fmt"
	"time"

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
	mc.MultiStatements = false
	mc.Loc = time.Local
	mc.Params = map[string]string{"charset": "utf8mb4"}
	mc.Collation = "utf8mb4_general_ci"

	db, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
