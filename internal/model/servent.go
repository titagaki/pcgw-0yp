package model

import (
	"database/sql"
)

type Servent struct {
	ID          int64
	Name        string
	Description string
	Hostname    string
	Port        int
	AuthID      string
	Passwd      string
	Priority    int
	MaxChannels int
	Enabled     bool
	Agent       string
	YellowPages string
}

func GetServent(db *sql.DB, id int64) (*Servent, error) {
	s := &Servent{}
	err := db.QueryRow(
		`SELECT id, name, description, hostname, port, auth_id, passwd, priority, max_channels, enabled, agent, yellow_pages FROM servents WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.Description, &s.Hostname, &s.Port, &s.AuthID, &s.Passwd, &s.Priority, &s.MaxChannels, &s.Enabled, &s.Agent, &s.YellowPages)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func ListServents(db *sql.DB) ([]*Servent, error) {
	rows, err := db.Query(
		`SELECT id, name, description, hostname, port, auth_id, passwd, priority, max_channels, enabled, agent, yellow_pages FROM servents ORDER BY priority, id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var servents []*Servent
	for rows.Next() {
		s := &Servent{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Hostname, &s.Port, &s.AuthID, &s.Passwd, &s.Priority, &s.MaxChannels, &s.Enabled, &s.Agent, &s.YellowPages); err != nil {
			return nil, err
		}
		servents = append(servents, s)
	}
	return servents, rows.Err()
}

func ListEnabledServents(db *sql.DB) ([]*Servent, error) {
	rows, err := db.Query(
		`SELECT id, name, description, hostname, port, auth_id, passwd, priority, max_channels, enabled, agent, yellow_pages FROM servents WHERE enabled = 1 ORDER BY priority, id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var servents []*Servent
	for rows.Next() {
		s := &Servent{}
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Hostname, &s.Port, &s.AuthID, &s.Passwd, &s.Priority, &s.MaxChannels, &s.Enabled, &s.Agent, &s.YellowPages); err != nil {
			return nil, err
		}
		servents = append(servents, s)
	}
	return servents, rows.Err()
}

// RequestServentWithVacancy returns an enabled servent with available capacity.
func RequestServentWithVacancy(db *sql.DB) (*Servent, error) {
	rows, err := ListEnabledServents(db)
	if err != nil {
		return nil, err
	}
	for _, s := range rows {
		count, err := CountChannelsByServent(db, s.ID)
		if err != nil {
			return nil, err
		}
		if s.MaxChannels == 0 || count < s.MaxChannels {
			return s, nil
		}
	}
	return nil, sql.ErrNoRows
}

func CreateServent(db *sql.DB, name, description, hostname string, port int, authID, passwd string, priority, maxChannels int, enabled bool) (*Servent, error) {
	result, err := db.Exec(
		`INSERT INTO servents (name, description, hostname, port, auth_id, passwd, priority, max_channels, enabled) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, description, hostname, port, authID, passwd, priority, maxChannels, enabled,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return GetServent(db, id)
}

func UpdateServent(db *sql.DB, id int64, name, description, hostname string, port int, authID, passwd string, priority, maxChannels int, enabled bool) error {
	_, err := db.Exec(
		`UPDATE servents SET name=?, description=?, hostname=?, port=?, auth_id=?, passwd=?, priority=?, max_channels=?, enabled=? WHERE id=?`,
		name, description, hostname, port, authID, passwd, priority, maxChannels, enabled, id,
	)
	return err
}

func UpdateServentAgent(db *sql.DB, id int64, agent, yellowPages string) error {
	_, err := db.Exec(`UPDATE servents SET agent=?, yellow_pages=? WHERE id=?`, agent, yellowPages, id)
	return err
}

func DeleteServent(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM servents WHERE id = ?`, id)
	return err
}
