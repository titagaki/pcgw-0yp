package domain

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int64
	Name            string
	Image           string
	TwitterID       sql.NullString
	Admin           bool
	Suspended       bool
	Bio             string
	NoticeCheckedAt sql.NullTime
	LoggedOnAt      sql.NullTime
	CreatedAt       time.Time
}
