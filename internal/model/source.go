package model

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

type Source struct {
	ID     int64
	UserID int64
	Name   string
	Key    string
}

func ListSourcesByUser(db *sql.DB, userID int64) ([]*Source, error) {
	rows, err := db.Query(`SELECT id, user_id, name, ` + "`key`" + ` FROM sources WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []*Source
	for rows.Next() {
		s := &Source{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.Key); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func GetSource(db *sql.DB, id int64) (*Source, error) {
	s := &Source{}
	err := db.QueryRow(`SELECT id, user_id, name, ` + "`key`" + ` FROM sources WHERE id = ?`, id).Scan(&s.ID, &s.UserID, &s.Name, &s.Key)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func generateKey() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func CreateSource(db *sql.DB, userID int64, name string) (*Source, error) {
	key := generateKey()
	result, err := db.Exec(`INSERT INTO sources (user_id, name, ` + "`key`" + `) VALUES (?, ?, ?)`, userID, name, key)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetSource(db, id)
}

func RegenerateSourceKey(db *sql.DB, id int64) error {
	key := generateKey()
	_, err := db.Exec(`UPDATE sources SET ` + "`key`" + ` = ? WHERE id = ?`, key, id)
	return err
}

func DeleteSource(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM sources WHERE id = ?`, id)
	return err
}

func CountSourcesByUser(db *sql.DB, userID int64) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM sources WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}
