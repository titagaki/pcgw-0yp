package domain

import (
	"database/sql"
	"time"
)

type Channel struct {
	ID           int64
	GnuID        string
	UserID       int64
	ServentID    int64
	StreamKey    string
	LastActiveAt sql.NullTime
	CreatedAt    time.Time
}
