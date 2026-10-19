package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func setupDatabase() error {
	var err error

	db, err = sql.Open(
		"sqlite",
		"file:researchbridge.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)",
	)

	if err != nil {
		return err
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()

	if err != nil {
		db.Close()
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL COLLATE NOCASE UNIQUE,
		password_hash TEXT NOT NULL,
		department TEXT NOT NULL DEFAULT '',
		batch TEXT NOT NULL DEFAULT '',
		skills TEXT NOT NULL DEFAULT '',
		research_interests TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(query)

	if err != nil {
		db.Close()
		return err
	}

	fmt.Println("SQLite database is ready")

	return nil
}