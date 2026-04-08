package model

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

func GetUser(db *sql.DB, id int64) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, name, image, twitter_id, admin, suspended, bio, notice_checked_at, logged_on_at, created_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin, &u.Suspended, &u.Bio, &u.NoticeCheckedAt, &u.LoggedOnAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func GetUserByTwitterID(db *sql.DB, twitterID string) (*User, error) {
	u := &User{}
	err := db.QueryRow(
		`SELECT id, name, image, twitter_id, admin, suspended, bio, notice_checked_at, logged_on_at, created_at FROM users WHERE twitter_id = ?`, twitterID,
	).Scan(&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin, &u.Suspended, &u.Bio, &u.NoticeCheckedAt, &u.LoggedOnAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func CreateUser(db *sql.DB, name, image, twitterID string) (*User, error) {
	result, err := db.Exec(
		`INSERT INTO users (name, image, twitter_id) VALUES (?, ?, ?)`,
		name, image, twitterID,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetUser(db, id)
}

func UpdateUser(db *sql.DB, id int64, name, image, bio string) error {
	_, err := db.Exec(
		`UPDATE users SET name = ?, image = ?, bio = ? WHERE id = ?`,
		name, image, bio, id,
	)
	return err
}

func UpdateUserLoggedOn(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE users SET logged_on_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func UpdateUserAdmin(db *sql.DB, id int64, admin bool) error {
	_, err := db.Exec(`UPDATE users SET admin = ? WHERE id = ?`, admin, id)
	return err
}

func UpdateUserSuspended(db *sql.DB, id int64, suspended bool) error {
	_, err := db.Exec(`UPDATE users SET suspended = ? WHERE id = ?`, suspended, id)
	return err
}

func UpdateUserNoticeChecked(db *sql.DB, id int64) error {
	_, err := db.Exec(`UPDATE users SET notice_checked_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

func ListUsers(db *sql.DB) ([]*User, error) {
	rows, err := db.Query(
		`SELECT id, name, image, twitter_id, admin, suspended, bio, notice_checked_at, logged_on_at, created_at FROM users ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin, &u.Suspended, &u.Bio, &u.NoticeCheckedAt, &u.LoggedOnAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func ListActiveUsers(db *sql.DB, days int) ([]*User, error) {
	since := time.Now().AddDate(0, 0, -days)
	rows, err := db.Query(
		`SELECT u.id, u.name, u.image, u.twitter_id, u.admin, u.suspended, u.bio, u.notice_checked_at, u.logged_on_at, u.created_at
		 FROM users u
		 WHERE u.logged_on_at >= ?
		   AND EXISTS (SELECT 1 FROM channel_infos ci WHERE ci.user_id = u.id)
		 ORDER BY u.logged_on_at DESC`, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin, &u.Suspended, &u.Bio, &u.NoticeCheckedAt, &u.LoggedOnAt, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func SearchUsersByName(db *sql.DB, query string) ([]*User, error) {
	rows, err := db.Query(
		`SELECT id, name, image, twitter_id, admin, suspended, bio, notice_checked_at, logged_on_at, created_at FROM users WHERE name LIKE ? ORDER BY name`, "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Image, &u.TwitterID, &u.Admin, &u.Suspended, &u.Bio, &u.NoticeCheckedAt, &u.LoggedOnAt, &u.CreatedAt); err != nil {
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
