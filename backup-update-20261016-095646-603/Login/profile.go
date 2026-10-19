package main

import (
	"database/sql"
	"errors"
	"strings"
	"unicode/utf8"
)

// Adds profile columns in place so existing accounts, tokens and proposals survive upgrades.
// Re-running startup is safe: only missing columns are added, within one transaction.
func migrateUserProfile(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	rows, err := tx.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notNull, primary int
		var name, typ string
		var defaultValue any
		if err = rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primary); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err = rows.Close(); err != nil {
		return err
	}
	columns := []struct{ name, definition string }{
		{"institution", "TEXT NOT NULL DEFAULT ''"}, {"department", "TEXT NOT NULL DEFAULT ''"},
		{"academic_year", "INTEGER NOT NULL DEFAULT 0"}, {"semester", "INTEGER NOT NULL DEFAULT 0"},
	}
	for _, column := range columns {
		if !existing[column.name] {
			if _, err = tx.Exec("ALTER TABLE users ADD COLUMN " + column.name + " " + column.definition); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
func validateProfile(c *credentials) error {
	c.Institution = strings.TrimSpace(c.Institution)
	c.Department = strings.TrimSpace(c.Department)
	if !utf8.ValidString(c.Institution) || utf8.RuneCountInString(c.Institution) < 2 || utf8.RuneCountInString(c.Institution) > 180 {
		return errors.New("Enter your institution name (2–180 characters).")
	}
	if c.Role == "teacher" {
		c.Department = ""
		c.AcademicYear = 0
		c.Semester = 0
		return nil
	}
	if !utf8.ValidString(c.Department) || utf8.RuneCountInString(c.Department) < 2 || utf8.RuneCountInString(c.Department) > 120 {
		return errors.New("Enter your department (2–120 characters).")
	}
	if c.AcademicYear < 1 || c.AcademicYear > 6 {
		return errors.New("Select your academic year (1–6).")
	}
	if c.Semester < 1 || c.Semester > 12 {
		return errors.New("Select your semester (1–12).")
	}
	return nil
}
