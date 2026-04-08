package repository

import (
	"database/sql"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/domain"
)

func GetNotice(db *sql.DB, id int64) (*domain.Notice, error) {
	n := &domain.Notice{}
	err := db.QueryRow(`
		SELECT id, title, body, created_at, updated_at
		FROM notices
		WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func ListNotices(db *sql.DB) ([]*domain.Notice, error) {
	rows, err := db.Query(`
		SELECT id, title, body, created_at, updated_at
		FROM notices
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notices []*domain.Notice
	for rows.Next() {
		n := &domain.Notice{}
		if err := rows.Scan(&n.ID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notices = append(notices, n)
	}
	return notices, rows.Err()
}

func HasUnreadNotices(db *sql.DB, userID int64) (bool, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM notices n, users u
		WHERE u.id = ?
		  AND (u.notice_checked_at IS NULL OR n.created_at > u.notice_checked_at)`,
		userID,
	).Scan(&count)
	return count > 0, err
}

func CreateNotice(db *sql.DB, title, body string) (*domain.Notice, error) {
	now := time.Now()
	result, err := db.Exec(`
		INSERT INTO notices (title, body, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		title, body, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetNotice(db, id)
}

func UpdateNotice(db *sql.DB, id int64, title, body string) error {
	_, err := db.Exec(`
		UPDATE notices
		SET title = ?, body = ?, updated_at = ?
		WHERE id = ?`,
		title, body, time.Now(), id,
	)
	return err
}

func DeleteNotice(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM notices WHERE id = ?`, id)
	return err
}
