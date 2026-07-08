package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// ConnectDB initialises the database connection and runs migrations
func ConnectDB(cfg *Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	switch cfg.DBType {
	case "postgres":
		db, err = sql.Open("postgres", cfg.DBURL)
		if err != nil {
			return nil, fmt.Errorf("failed to open postgres: %w", err)
		}
		db.SetMaxOpenConns(100)
		db.SetMaxIdleConns(10)

	default: // sqlite
		// Ensure directory exists
		dir := filepath.Dir(cfg.DBURL)
		if dir != "." && dir != "" {
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				return nil, fmt.Errorf("failed to create data dir: %w", mkErr)
			}
		}
		db, err = sql.Open("sqlite3", cfg.DBURL+"?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on")
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite: %w", err)
		}
		// SQLite performs best with a single writer
		db.SetMaxOpenConns(1)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connected (%s)", cfg.DBType)
	DB = db
	return db, nil
}
