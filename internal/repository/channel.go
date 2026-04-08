package repository

import (
	"database/sql"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/domain"
)

const channelColumns = `
	id, gnu_id, user_id, servent_id,
	stream_key, last_active_at, created_at`

func GetChannel(db *sql.DB, id int64) (*domain.Channel, error) {
	ch := &domain.Channel{}
	err := db.QueryRow(`
		SELECT`+channelColumns+`
		FROM channels
		WHERE id = ?`, id,
	).Scan(
		&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID,
		&ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func GetChannelByGnuID(db *sql.DB, gnuID string) (*domain.Channel, error) {
	ch := &domain.Channel{}
	err := db.QueryRow(`
		SELECT`+channelColumns+`
		FROM channels
		WHERE gnu_id = ?`, gnuID,
	).Scan(
		&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID,
		&ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func scanChannel(rows *sql.Rows) (*domain.Channel, error) {
	ch := &domain.Channel{}
	if err := rows.Scan(
		&ch.ID, &ch.GnuID, &ch.UserID, &ch.ServentID,
		&ch.StreamKey, &ch.LastActiveAt, &ch.CreatedAt,
	); err != nil {
		return nil, err
	}
	return ch, nil
}

func ListChannels(db *sql.DB) ([]*domain.Channel, error) {
	rows, err := db.Query(`
		SELECT` + channelColumns + `
		FROM channels
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []*domain.Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func ListChannelsByUser(db *sql.DB, userID int64) ([]*domain.Channel, error) {
	rows, err := db.Query(`
		SELECT`+channelColumns+`
		FROM channels
		WHERE user_id = ?
		ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var channels []*domain.Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, rows.Err()
}

func CountChannelsByServent(db *sql.DB, serventID int64) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM channels WHERE servent_id = ?`, serventID,
	).Scan(&count)
	return count, err
}

func CreateChannel(db *sql.DB, gnuID string, userID, serventID int64, streamKey string) (*domain.Channel, error) {
	result, err := db.Exec(`
		INSERT INTO channels (gnu_id, user_id, servent_id, stream_key)
		VALUES (?, ?, ?, ?)`,
		gnuID, userID, serventID, streamKey,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetChannel(db, id)
}

func UpdateChannelLastActive(db *sql.DB, id int64) error {
	_, err := db.Exec(
		`UPDATE channels SET last_active_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func DeleteChannel(db *sql.DB, id int64) error {
	db.Exec(
		`UPDATE channel_infos SET channel_id = NULL, terminated_at = ? WHERE channel_id = ?`,
		time.Now(), id,
	)
	_, err := db.Exec(`DELETE FROM channels WHERE id = ?`, id)
	return err
}
