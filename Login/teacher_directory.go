package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
}

func (a *App) teacherDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if _, err := a.user(r); err != nil {
		fail(w, http.StatusUnauthorized, "Please log in.")
		return
	}
	if !a.db.postgres {
		fail(w, http.StatusServiceUnavailable, "Teacher directory requires the shared PostgreSQL database. Configure DATABASE_URL.")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	// The configured PostgreSQL search_path points to the application's schema.
	rows, err := a.db.DB.QueryContext(ctx, `SELECT id,name,designation,institution,department,phone,email,website_url,photo_url,research_interests,expertise,accepting_students,source,verified FROM teacher_directory ORDER BY name,id`)
	if err != nil {
		log.Printf("Teacher directory query: %v", err)
		fail(w, 500, "Teacher directory unavailable. Check the database connection and teacher_directory migration.")
		return
	}
	defer rows.Close()
	entries, err := readTeacherDirectory(rows)
	if err != nil {
		log.Printf("Teacher directory decode: %v", err)
		fail(w, 500, "Teacher directory unavailable.")
		return
	}
	reply(w, 200, entries)
}

func readTeacherDirectory(rows *sql.Rows) ([]teacherDirectoryEntry, error) {
	entries := make([]teacherDirectoryEntry, 0)
	for rows.Next() {
		var item teacherDirectoryEntry
		var interests, expertise []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Designation, &item.Institution, &item.Department, &item.Phone, &item.Email, &item.WebsiteURL, &item.PhotoURL, &interests, &expertise, &item.AcceptingStudents, &item.Source, &item.Verified); err != nil {
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
