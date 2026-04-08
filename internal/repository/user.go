package repository

import (
	"database/sql"
	"time"

	"github.com/titagaki/pcgw-0yp/internal/domain"
)

const userColumns = `
	id, name, image, twitter_id, admin,
	suspended, bio, notice_checked_at,
	logged_on_at, created_at`

func scanUser(row interface{ Scan(...interface{}) error }) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(
		&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin,
		&u.Suspended, &u.Bio, &u.NoticeCheckedAt,
		&u.LoggedOnAt, &u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUser(db *sql.DB, id int64) (*domain.User, error) {
	return scanUser(db.QueryRow(`
		SELECT`+userColumns+`
		FROM users
		WHERE id = ?`, id))
}

func GetUserByTwitterID(db *sql.DB, twitterID string) (*domain.User, error) {
	return scanUser(db.QueryRow(`
		SELECT`+userColumns+`
		FROM users
		WHERE twitter_id = ?`, twitterID))
}

func CreateUser(db *sql.DB, name, image, twitterID string) (*domain.User, error) {
	result, err := db.Exec(`
		INSERT INTO users (name, image, twitter_id)
		VALUES (?, ?, ?)`,
		name, image, twitterID,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetUser(db, id)
}

func UpdateUser(db *sql.DB, id int64, name, image, bio string) error {
	_, err := db.Exec(`
		UPDATE users
		SET name = ?, image = ?, bio = ?
		WHERE id = ?`,
		name, image, bio, id,
	)
	return err
}

func UpdateUserLoggedOn(db *sql.DB, id int64) error {
	_, err := db.Exec(
		`UPDATE users SET logged_on_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func UpdateUserAdmin(db *sql.DB, id int64, admin bool) error {
	_, err := db.Exec(
		`UPDATE users SET admin = ? WHERE id = ?`,
		admin, id,
	)
	return err
}

func UpdateUserSuspended(db *sql.DB, id int64, suspended bool) error {
	_, err := db.Exec(
		`UPDATE users SET suspended = ? WHERE id = ?`,
		suspended, id,
	)
	return err
}

func UpdateUserNoticeChecked(db *sql.DB, id int64) error {
	_, err := db.Exec(
		`UPDATE users SET notice_checked_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

func ListUsers(db *sql.DB) ([]*domain.User, error) {
	rows, err := db.Query(`
		SELECT` + userColumns + `
		FROM users
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func ListActiveUsers(db *sql.DB, days int) ([]*domain.User, error) {
	since := time.Now().AddDate(0, 0, -days)
	rows, err := db.Query(`
		SELECT
			u.id, u.name, u.image, u.twitter_id, u.admin,
			u.suspended, u.bio, u.notice_checked_at,
			u.logged_on_at, u.created_at
		FROM users u
		WHERE u.logged_on_at >= ?
		  AND EXISTS (SELECT 1 FROM channel_infos ci WHERE ci.user_id = u.id)
		ORDER BY u.logged_on_at DESC`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func SearchUsersByName(db *sql.DB, query string) ([]*domain.User, error) {
	rows, err := db.Query(`
		SELECT`+userColumns+`
		FROM users
		WHERE name LIKE ?
		ORDER BY name`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func DeleteUser(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}
