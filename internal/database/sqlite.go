package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// User represents a user in the system
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Plan         string    `json:"plan"`
	CreatedAt    time.Time `json:"created_at"`
}

// DB is the global database connection
var DB *sql.DB

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	fmt.Printf("📦 Database initialized: %s\n", dbPath)
	return nil
}

// createTables creates the necessary database tables
func createTables() error {
	// Users table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			plan TEXT DEFAULT 'free',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Metrics table for analytics
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			backend_url TEXT,
			response_time_ms INTEGER,
			status_code INTEGER,
			is_error BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// Create index for faster queries
	_, err = DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)
	`)

	return err
}

// CreateUser creates a new user in the database
func CreateUser(email, passwordHash string) (*User, error) {
	result, err := DB.Exec(
		"INSERT INTO users (email, password_hash, plan) VALUES (?, ?, 'free')",
		email, passwordHash,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	id, _ := result.LastInsertId()
	return &User{
		ID:        id,
		Email:     email,
		Plan:      "free",
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, email, password_hash, plan, created_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Plan, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int64) (*User, error) {
	user := &User{}
	err := DB.QueryRow(
		"SELECT id, email, password_hash, plan, created_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Plan, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserPlan updates a user's subscription plan
func UpdateUserPlan(userID int64, plan string) error {
	_, err := DB.Exec("UPDATE users SET plan = ? WHERE id = ?", plan, userID)
	return err
}

// Close closes the database connection
func Close() {
	if DB != nil {
		DB.Close()
	}
}
