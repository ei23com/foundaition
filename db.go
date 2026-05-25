package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// ─── Database Connection ─────────────────────────────────────────────────────

// openSQLite opens (or creates) a SQLite database file using modernc.org/sqlite.
func openSQLite(filePath string) (*sql.DB, error) {
	if !filepath.IsAbs(filePath) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, fmt.Errorf("resolve SQLite path: %w", err)
		}
		filePath = absPath
	}

	dsn := filePath + "?_timeout=5000&_journal_mode=WAL"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open SQLite: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping SQLite: %w", err)
	}

	log.Printf("Connected to SQLite: %s", filePath)
	return db, nil
}

// ─── Schema Management ───────────────────────────────────────────────────────

// ensureTableExists creates the links table if it does not exist.
// All column names are English for consistency.
func ensureTableExists(db *sql.DB, tblName string) error {
	createSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp    TEXT NOT NULL,
			url          TEXT NOT NULL,
			title        TEXT,
			note         TEXT,
			summary      TEXT,
			content      TEXT,
			category     TEXT,
			included     INTEGER DEFAULT 0,
			marked       INTEGER DEFAULT 0,
			read         INTEGER DEFAULT 0
		)`, tblName)

	_, err := db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("ensure table %s: %w", tblName, err)
	}

	var cnt int
	querySQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s LIMIT 1`, tblName)
	if err := db.QueryRow(querySQL).Scan(&cnt); err != nil {
		return fmt.Errorf("verify table %s: %w", tblName, err)
	}

	log.Printf("Table %s ready (%d entries)", tblName, cnt)
	return nil
}
