package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type teacherDirectoryEntry struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Designation       string   `json:"designation"`
	Institution       string   `json:"institution"`
	Department        string   `json:"department"`
	Phone             *string  `json:"phone"`
	Email             string   `json:"email"`
	WebsiteURL        *string  `json:"website_url"`
	PhotoURL          *string  `json:"photo_url"`
	ResearchInterests []string `json:"research_interests"`
	Expertise         []string `json:"expertise"`
	AcceptingStudents *bool    `json:"accepting_students"`
	Source            string   `json:"source"`
	Verified          bool     `json:"verified"`

	// This is the ID from the application's users table,
	// not the external teacher_directory record ID.
	// It remains nil when no verified teacher account matches.
	AccountID *int64 `json:"account_id"`
}

func (a *App) teacherDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	if _, err := a.user(r); err != nil {
		fail(w, http.StatusUnauthorized, "Please log in.")
		return
	}

	if !a.db.postgres {
		fail(
			w,
			http.StatusServiceUnavailable,
			"Teacher directory requires the shared PostgreSQL database. Configure DATABASE_URL.",
		)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Read every field from the external teacher directory.
	rows, err := a.db.DB.QueryContext(ctx, `
		SELECT
			id,
			name,
			designation,
			institution,
			department,
			phone,
			email,
			website_url,
			photo_url,
			research_interests,
			expertise,
			accepting_students,
			source,
			verified
		FROM teacher_directory
		ORDER BY name, id
	`)
	if err != nil {
		log.Printf("Teacher directory query failed: %v", err)
		fail(
			w,
			http.StatusInternalServerError,
			"Teacher directory unavailable. Check the database connection and teacher_directory table.",
		)
		return
	}

	entries, err := readTeacherDirectory(rows)
	closeErr := rows.Close()

	if err != nil {
		log.Printf("Teacher directory decode failed: %v", err)
		fail(w, http.StatusInternalServerError, "Teacher directory could not be read.")
		return
	}
	if closeErr != nil {
		log.Printf("Teacher directory close failed: %v", closeErr)
		fail(w, http.StatusInternalServerError, "Teacher directory could not be read.")
		return
	}

	// Match external directory emails with verified app teacher accounts.
	// A directory ID is not a users.id and must never be used as teacher_id.
	accountRows, err := a.db.DB.QueryContext(ctx, `
		SELECT id, email
		FROM users
		WHERE role = 'teacher' AND verified = 1
	`)
	if err != nil {
		log.Printf("Teacher account lookup failed: %v", err)
		fail(w, http.StatusInternalServerError, "Registered teacher accounts could not be checked.")
		return
	}

	if err := attachRegisteredTeacherAccounts(entries, accountRows); err != nil {
		log.Printf("Teacher account mapping failed: %v", err)
		fail(w, http.StatusInternalServerError, "Registered teacher accounts could not be checked.")
		return
	}
	// Add verified ResearchBridge teacher accounts that do not yet
	// have a matching row in teacher_directory.
	appRows, err := a.db.DB.QueryContext(ctx, `
    SELECT
        u.id,
        u.name,
        u.email,
        COALESCE(u.institution, ''),
        COALESCE(u.department, ''),
        COALESCE(rp.qualification, ''),
        COALESCE(rp.interests, ''),
        COALESCE(rp.expertise, '')
    FROM users u
    LEFT JOIN researcher_profiles rp ON rp.user_id = u.id
    WHERE u.role = 'teacher' AND u.verified = 1
    ORDER BY u.name
`)
	if err != nil {
		log.Printf("Registered teacher query failed: %v", err)
		fail(w, http.StatusInternalServerError, "Registered teachers could not be loaded.")
		return
	}

	

	entries, err = appendRegisteredTeachers(entries, appRows)
	if err != nil {

		log.Printf("Registered teacher merge failed: %v", err)
		fail(w, http.StatusInternalServerError, "Registered teachers could not be added.")
		return
	}
	reply(w, http.StatusOK, entries)
}

func readTeacherDirectory(rows *sql.Rows) ([]teacherDirectoryEntry, error) {
	entries := make([]teacherDirectoryEntry, 0)

	for rows.Next() {
		var item teacherDirectoryEntry
		var interests, expertise []byte

		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Designation,
			&item.Institution,
			&item.Department,
			&item.Phone,
			&item.Email,
			&item.WebsiteURL,
			&item.PhotoURL,
			&interests,
			&expertise,
			&item.AcceptingStudents,
			&item.Source,
			&item.Verified,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(interests, &item.ResearchInterests); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(expertise, &item.Expertise); err != nil {
			return nil, err
		}

		if item.ResearchInterests == nil {
			item.ResearchInterests = []string{}
		}
		if item.Expertise == nil {
			item.Expertise = []string{}
		}

		entries = append(entries, item)
	}

	return entries, rows.Err()
}

// attachRegisteredTeacherAccounts links directory entries to real app accounts.
// Entries without a match remain visible but cannot send a supervision request.
func attachRegisteredTeacherAccounts(
	entries []teacherDirectoryEntry,
	rows *sql.Rows,
) error {
	defer rows.Close()

	accountIDs := make(map[string]int64)

	for rows.Next() {
		var id int64
		var email string

		if err := rows.Scan(&id, &email); err != nil {
			return err
		}

		key := strings.ToLower(strings.TrimSpace(email))
		if key != "" {
			accountIDs[key] = id
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	for i := range entries {
		entries[i].AccountID = nil

		key := strings.ToLower(strings.TrimSpace(entries[i].Email))
		if id, ok := accountIDs[key]; ok && key != "" {
			matchedID := id
			entries[i].AccountID = &matchedID
		}
	}

	return nil
}
func appendRegisteredTeachers(
	entries []teacherDirectoryEntry,
	rows *sql.Rows,
) ([]teacherDirectoryEntry, error) {
	defer rows.Close()

	// Keep track of directory emails to avoid duplicate teacher cards.
	existingEmails := make(map[string]bool)

	for _, entry := range entries {
		key := strings.ToLower(strings.TrimSpace(entry.Email))
		if key != "" {
			existingEmails[key] = true
		}
	}

	for rows.Next() {
		var (
			id                       int64
			name, email              string
			institution, department  string
			qualification, interests string
			expertise                string
		)

		if err := rows.Scan(
			&id,
			&name,
			&email,
			&institution,
			&department,
			&qualification,
			&interests,
			&expertise,
		); err != nil {
			return nil, err
		}

		key := strings.ToLower(strings.TrimSpace(email))

		// This teacher already has a directory profile.
		if key != "" && existingEmails[key] {
			continue
		}

		accountID := id

		entries = append(entries, teacherDirectoryEntry{
			ID:                "account-" + strconv.FormatInt(id, 10),
			Name:              name,
			Designation:       qualification,
			Institution:       institution,
			Department:        department,
			Email:             email,
			ResearchInterests: splitTeacherProfileList(interests),
			Expertise:         splitTeacherProfileList(expertise),
			Source:            "ResearchBridge account",
			Verified:          true,
			AccountID:         &accountID,
		})

		if key != "" {
			existingEmails[key] = true
		}
	}

	return entries, rows.Err()
}

func splitTeacherProfileList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '|'
	})

	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
