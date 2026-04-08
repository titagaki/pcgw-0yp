package repository

import (
	"database/sql"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/domain"
)

const channelInfoColumns = `
	id, user_id, channel, genre, description,
	comment, url, stream_type, yp,
	channel_id, servent_id, source_name,
	terminated_at, created_at, updated_at`

func scanChannelInfo(row interface{ Scan(...interface{}) error }) (*domain.ChannelInfo, error) {
	ci := &domain.ChannelInfo{}
	err := row.Scan(
		&ci.ID, &ci.UserID, &ci.Channel, &ci.Genre, &ci.Description,
		&ci.Comment, &ci.URL, &ci.StreamType, &ci.YP,
		&ci.ChannelID, &ci.ServentID, &ci.SourceName,
		&ci.TerminatedAt, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ci, nil
}

func GetChannelInfo(db *sql.DB, id int64) (*domain.ChannelInfo, error) {
	return scanChannelInfo(db.QueryRow(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		WHERE id = ?`, id))
}

func GetChannelInfoByChannelID(db *sql.DB, channelID int64) (*domain.ChannelInfo, error) {
	return scanChannelInfo(db.QueryRow(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		WHERE channel_id = ?
		ORDER BY id DESC
		LIMIT 1`, channelID))
}

func GetLatestChannelInfoByUser(db *sql.DB, userID int64) (*domain.ChannelInfo, error) {
	return scanChannelInfo(db.QueryRow(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		WHERE user_id = ?
		ORDER BY id DESC
		LIMIT 1`, userID))
}

func ListChannelInfosByUser(db *sql.DB, userID int64, limit int) ([]*domain.ChannelInfo, error) {
	rows, err := db.Query(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var infos []*domain.ChannelInfo
	for rows.Next() {
		ci, err := scanChannelInfo(rows)
		if err != nil {
			return nil, err
		}
		infos = append(infos, ci)
	}
	return infos, rows.Err()
}

func ListRecentChannelInfos(db *sql.DB, limit int) ([]*domain.ChannelInfo, error) {
	rows, err := db.Query(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		ORDER BY created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var infos []*domain.ChannelInfo
	for rows.Next() {
		ci, err := scanChannelInfo(rows)
		if err != nil {
			return nil, err
		}
		infos = append(infos, ci)
	}
	return infos, rows.Err()
}

func ListChannelInfosByMonth(db *sql.DB, year, month int) ([]*domain.ChannelInfo, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	rows, err := db.Query(`
		SELECT`+channelInfoColumns+`
		FROM channel_infos
		WHERE created_at >= ? AND created_at < ?
		ORDER BY created_at DESC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var infos []*domain.ChannelInfo
	for rows.Next() {
		ci, err := scanChannelInfo(rows)
		if err != nil {
			return nil, err
		}
		infos = append(infos, ci)
	}
	return infos, rows.Err()
}

func CreateChannelInfo(db *sql.DB, userID int64, channel, genre, desc, comment, url, streamType, yp string, channelID, serventID sql.NullInt64, sourceName string) (*domain.ChannelInfo, error) {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO channel_infos
			(user_id, channel, genre, description, comment,
			 url, stream_type, yp, channel_id, servent_id,
			 source_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, channel, genre, desc, comment,
		url, streamType, yp, channelID, serventID,
		sourceName, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetChannelInfo(db, id)
}

func UpdateChannelInfo(db *sql.DB, id int64, channel, genre, desc, comment, url string) error {
	_, err := db.Exec(`
		UPDATE channel_infos
		SET channel = ?, genre = ?, description = ?,
		    comment = ?, url = ?, updated_at = ?
		WHERE id = ?`,
		channel, genre, desc, comment, url, time.Now(), id,
	)
	return err
}

func TerminateChannelInfo(db *sql.DB, channelID int64) error {
	_, err := db.Exec(`
		UPDATE channel_infos
		SET channel_id = NULL, terminated_at = ?
		WHERE channel_id = ?`,
		time.Now(), channelID,
	)
	return err
}

func DeleteChannelInfo(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM channel_infos WHERE id = ?`, id)
	return err
}
