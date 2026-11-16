package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type academicRecordInput struct {
	Year int     `json:"year"`
	CGPA float64 `json:"cgpa"`
}
type academicProfileInput struct {
	Records   []academicRecordInput `json:"records"`
	Skills    []string              `json:"skills"`
	Interests string                `json:"interests"`
}

// migrateAcademicProfile creates private student academic data tables.
// The same SQL works on SQLite and PostgreSQL.
func migrateAcademicProfile(db *Database) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS student_academic_records (
		id INTEGER PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		academic_year INTEGER NOT NULL CHECK(academic_year BETWEEN 1 AND 6),
		cgpa REAL NOT NULL CHECK(cgpa >= 0 AND cgpa <= 4),
		updated_at BIGINT NOT NULL,
		UNIQUE(user_id, academic_year)
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS student_skills (
		id INTEGER PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		skill TEXT NOT NULL,
		UNIQUE(user_id, skill)
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS student_academic_interests (
		user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
		interests TEXT NOT NULL DEFAULT '',
		updated_at BIGINT NOT NULL
	)`)
	return err
}

// GET/PUT /api/student-academic-profile: only the logged-in student can access/edit their own record.
func (a *App) studentAcademicProfile(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Student account required.")
		return
	}
	if r.Method == http.MethodGet {
		rows, err := a.db.Query(`SELECT academic_year,cgpa FROM student_academic_records WHERE user_id=? ORDER BY academic_year`, u.ID)
		if err != nil {
			fail(w, 500, "Could not load academic records.")
			return
		}
		records := []academicRecordInput{}
		for rows.Next() {
			var v academicRecordInput
			if rows.Scan(&v.Year, &v.CGPA) != nil {
				rows.Close()
				fail(w, 500, "Could not read academic records.")
				return
			}
			records = append(records, v)
		}
		if err = rows.Err(); err != nil {
			rows.Close()
			fail(w, 500, "Could not read academic records.")
			return
		}
		rows.Close()
		srows, err := a.db.Query(`SELECT skill FROM student_skills WHERE user_id=? ORDER BY skill`, u.ID)
		if err != nil {
			fail(w, 500, "Could not load skills.")
			return
		}
		skills := []string{}
		for srows.Next() {
			var v string
			if srows.Scan(&v) != nil {
				srows.Close()
				fail(w, 500, "Could not read skills.")
				return
			}
			skills = append(skills, v)
		}
		srows.Close()
		var interests string
		_ = a.db.QueryRow(`SELECT interests FROM student_academic_interests WHERE user_id=?`, u.ID).Scan(&interests)
		reply(w, 200, map[string]any{"records": records, "skills": skills, "interests": interests})
		return
	}
	if r.Method != http.MethodPut {
		fail(w, 405, "Method not allowed.")
		return
	}
	var in academicProfileInput
	if !decode(w, r, &in) {
		return
	}
	if len(in.Records) > 6 || len(in.Skills) > 30 || len(strings.TrimSpace(in.Interests)) > 2000 {
		fail(w, 400, "Too many records, skills or interests.")
		return
	}
	seen := map[int]bool{}
	for _, v := range in.Records {
		if v.Year < 1 || v.Year > 6 || seen[v.Year] || v.CGPA < 0 || v.CGPA > 4 {
			fail(w, 400, "Each academic year must be unique and CGPA must be between 0 and 4.")
			return
		}
		seen[v.Year] = true
	}
	clean := []string{}
	skillSeen := map[string]bool{}
	for _, v := range in.Skills {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if len(v) > 60 {
			fail(w, 400, "Each skill must be at most 60 characters.")
			return
		}
		key := strings.ToLower(v)
		if !skillSeen[key] {
			skillSeen[key] = true
			clean = append(clean, v)
		}
	}
	tx, err := a.db.Begin()
	if err != nil {
		fail(w, 500, "Could not save profile.")
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`DELETE FROM student_academic_records WHERE user_id=?`, u.ID); err != nil {
		fail(w, 500, "Could not save academic records.")
		return
	}
	for _, v := range in.Records {
		if _, err = tx.Exec(`INSERT INTO student_academic_records(user_id,academic_year,cgpa,updated_at) VALUES(?,?,?,?)`, u.ID, v.Year, v.CGPA, time.Now().Unix()); err != nil {
			fail(w, 500, "Could not save academic records.")
			return
		}
	}
	if _, err = tx.Exec(`DELETE FROM student_skills WHERE user_id=?`, u.ID); err != nil {
		fail(w, 500, "Could not save skills.")
		return
	}
	for _, v := range clean {
		if _, err = tx.Exec(`INSERT INTO student_skills(user_id,skill) VALUES(?,?)`, u.ID, v); err != nil {
			fail(w, 500, "Could not save skills.")
			return
		}
	}
	if _, err = tx.Exec(`INSERT INTO student_academic_interests(user_id,interests,updated_at) VALUES(?,?,?) ON CONFLICT(user_id) DO UPDATE SET interests=excluded.interests,updated_at=excluded.updated_at`, u.ID, strings.TrimSpace(in.Interests), time.Now().Unix()); err != nil {
		fail(w, 500, "Could not save interests.")
		return
	}
	if err = tx.Commit(); err != nil {
		fail(w, 500, "Could not save profile.")
		return
	}
	reply(w, 200, map[string]string{"message": "Academic profile saved."})
}

// GET /api/projects/{id}/team-academics is deliberately restricted to the teacher
// selected on that exact pending/accepted supervision request.
func (a *App) teamAcademicProfiles(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "teacher" {
		fail(w, 403, "Teacher account required.")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "Invalid project.")
		return
	}
	var owner int64
	var state string
	err = a.db.QueryRow(`SELECT teacher_id,COALESCE(supervision_status,'none') FROM projects WHERE id=?`, id).Scan(&owner, &state)
	if err != nil {
		fail(w, 404, "Request not found.")
		return
	}
	if owner != u.ID || (state != "pending" && state != "accepted" && state != "declined") {
		fail(w, 403, "You are not authorized to view this team's academic profiles.")
		return
	}
	rows, err := a.db.Query(`
		SELECT u.id, u.name, u.email
		FROM users u
		JOIN (
			SELECT student_id AS user_id
			FROM projects
			WHERE id = ?
			UNION
			SELECT user_id
			FROM team_members
			WHERE project_id = ?
		) team ON team.user_id = u.id
		ORDER BY u.name
	`, id, id)
	if err != nil {
		fail(w, 500, "Could not load team.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var uid int64
		var name, email string
		if rows.Scan(&uid, &name, &email) != nil {
			fail(w, 500, "Could not read team.")
			return
		}
		recRows, e := a.db.Query(`SELECT academic_year,cgpa FROM student_academic_records WHERE user_id=? ORDER BY academic_year`, uid)
		if e != nil {
			fail(w, 500, "Could not load academic records.")
			return
		}
		records := []map[string]any{}
		for recRows.Next() {
			var y int
			var cg float64
			if recRows.Scan(&y, &cg) != nil {
				recRows.Close()
				fail(w, 500, "Could not read academic records.")
				return
			}
			records = append(records, map[string]any{"year": y, "cgpa": cg})
		}
		recRows.Close()
		sr, e := a.db.Query(`SELECT skill FROM student_skills WHERE user_id=? ORDER BY skill`, uid)
		if e != nil {
			fail(w, 500, "Could not load skills.")
			return
		}
		skills := []string{}
		for sr.Next() {
			var v string
			if sr.Scan(&v) != nil {
				sr.Close()
				fail(w, 500, "Could not read skills.")
				return
			}
			skills = append(skills, v)
		}
		sr.Close()
		var interests string
		_ = a.db.QueryRow(`SELECT interests FROM student_academic_interests WHERE user_id=?`, uid).Scan(&interests)
		out = append(out, map[string]any{"user_id": uid, "name": name, "email": email, "records": records, "skills": skills, "interests": interests})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read team.")
		return
	}
	reply(w, 200, out)
}
