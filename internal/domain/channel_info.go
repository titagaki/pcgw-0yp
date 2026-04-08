package domain

import (
	"database/sql"
	"time"
)

type ChannelInfo struct {
	ID           int64
	UserID       int64
	Channel      string
	Genre        string
	Description  string
	Comment      string
	URL          string
	StreamType   string
	YP           string
	ChannelID    sql.NullInt64
	ServentID    sql.NullInt64
	SourceName   string
	TerminatedAt sql.NullTime
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (ci *ChannelInfo) IsActive() bool {
	return ci.ChannelID.Valid && !ci.TerminatedAt.Valid
}
