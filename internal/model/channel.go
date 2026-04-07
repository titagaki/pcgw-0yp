package model

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

func GetChannel(db *sql.DB, id int64) (*Channel, error) {
	ch := &Channel{}
	err := db.QueryRow(
		`SELECT id, gnu_id, user_id, servent_id, stream_key, last_active_at, created_at FROM channels WHERE id = ?`, id,
	).Scan(&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID, &ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func GetChannelByGnuID(db *sql.DB, gnuID string) (*Channel, error) {
	ch := &Channel{}
	err := db.QueryRow(
		`SELECT id, gnu_id, user_id, servent_id, stream_key, last_active_at, created_at FROM channels WHERE gnu_id = ?`, gnuID,
	).Scan(&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID, &ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func ListChannels(db *sql.DB) ([]*Channel, error) {
	rows, err := db.Query(
		`SELECT id, gnu_id, user_id, servent_id, stream_key, last_active_at, created_at FROM channels ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID, &ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func ListChannelsByUser(db *sql.DB, userID int64) ([]*Channel, error) {
	rows, err := db.Query(
		`SELECT id, gnu_id, user_id, servent_id, stream_key, last_active_at, created_at FROM channels WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []*Channel
	for rows.Next() {
		ch := &Channel{}
		if err := rows.Scan(&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID, &ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func CountChannelsByServent(db *sql.DB, serventID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM channels WHERE servent_id = ?`, serventID).Scan(&count)
	return count, err
}

func CreateChannel(db *sql.DB, gnuID string, userID, serventID int64, streamKey string) (*Channel, error) {
	result, err := db.Exec(
		`INSERT INTO channels (gnu_id, user_id, servent_id, stream_key) VALUES (?, ?, ?, ?)`,
		gnuID, userID, serventID, streamKey,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetChannel(db, id)
}

func UpdateChannelLastActive(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE channels SET last_active_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func DeleteChannel(db *sql.DB, id int64) error {
	// Clear reference from channel_infos first
	db.Exec(`UPDATE channel_infos SET channel_id = NULL, terminated_at = ? WHERE channel_id = ?`, time.Now(), id)
	_, err := db.Exec(`DELETE FROM channels WHERE id = ?`, id)
	return err
}
