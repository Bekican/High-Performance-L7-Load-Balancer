package store

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
)

var DB *sql.DB

func InitSQLite(filepath string) error {
	var err error
	DB, err = sql.Open("sqlite", filepath)
	if err != nil {
		return err
	}

	// Create tables if not exist
	query := `
	CREATE TABLE IF NOT EXISTS backends (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL UNIQUE,
		active BOOLEAN DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE,
		quota INTEGER DEFAULT 1000
	);
	`
	_, err = DB.Exec(query)
	if err != nil {
		return fmt.Errorf("tablolar oluşturulamadı: %v", err)
	}

	return nil
}

// AddBackend adds a new backend to the database
func AddBackend(url string) error {
	_, err := DB.Exec("INSERT INTO backends (url, active) VALUES (?, 1)", url)
	return err
}

// RemoveBackend removes a backend from the database
func RemoveBackend(url string) error {
	_, err := DB.Exec("DELETE FROM backends WHERE url = ?", url)
	return err
}

// GetBackends retrieves all active backends
func GetBackends() ([]string, error) {
	rows, err := DB.Query("SELECT url FROM backends WHERE active = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backends []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		backends = append(backends, url)
	}
	return backends, nil
}
