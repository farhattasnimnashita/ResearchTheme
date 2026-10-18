package main

import (
	"net/http"
	"strings"
	"time"
)

type DirectoryStudent struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Institution  string `json:"institution"`
	Department   string `json:"department"`
	AcademicYear int    `json:"academic_year"`
	Semester     int    `json:"semester"`
	Skills       string `json:"skills"`
	Interests    string `json:"interests"`
}

func (a *App) searchStudents(w http.ResponseWriter, r *http.Request) {
	me, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if me.Role != "student" {
		fail(w, 403, "Students only.")
		return
	}

	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) > 100 {
		fail(w, 400, "Search text is too long.")
		return
	}

	pattern := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(q) + "%"
	rows, err := a.db.Query(`
        SELECT u.id, u.name, u.institution,
               u.department, u.academic_year, u.semester,
               COALESCE(rp.expertise, ''),
               COALESCE(rp.interests, '')
        FROM users u
        LEFT JOIN researcher_profiles rp ON rp.user_id = u.id
        WHERE u.role = 'student'
          AND u.verified = 1
          AND u.id != ?
          AND (
              ? = ''
              OR lower(u.name) LIKE lower(?) ESCAPE '!'
              OR lower(u.department) LIKE lower(?) ESCAPE '!'
              OR lower(u.institution) LIKE lower(?) ESCAPE '!'
              OR lower(COALESCE(rp.expertise,'')) LIKE lower(?) ESCAPE '!'
              OR lower(COALESCE(rp.interests,'')) LIKE lower(?) ESCAPE '!'
          )
        ORDER BY lower(u.name),u.id
        LIMIT 50
    `, me.ID, q, pattern, pattern, pattern, pattern, pattern)

	if err != nil {
		fail(w, 500, "Could not search students.")
		return
	}
	defer rows.Close()

	students := []DirectoryStudent{}

	for rows.Next() {
		var s DirectoryStudent
		err := rows.Scan(
			&s.ID, &s.Name, &s.Institution,
			&s.Department, &s.AcademicYear,
			&s.Semester, &s.Skills, &s.Interests,
		)
		if err != nil {
			fail(w, 500, "Could not read students.")
			return
		}
		students = append(students, s)
	}

	if rows.Err() != nil {
		fail(w, 500, "Could not read students.")
		return
	}

	reply(w, 200, students)
}

func (a *App) inviteDirectoryStudent(w http.ResponseWriter, r *http.Request) {
	me, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if me.Role != "student" {
		fail(w, 403, "Students only.")
		return
	}

	var input struct {
		ProjectID int64 `json:"project_id"`
		StudentID int64 `json:"student_id"`
	}

	if !decode(w, r, &input) {
		return
	}

	if input.ProjectID <= 0 || input.StudentID <= 0 || input.StudentID == me.ID {
		fail(w, 400, "Invalid project or student.")
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		fail(w, 500, "Could not send invitation.")
		return
	}
	defer tx.Rollback()

	var ownerID int64
	err = tx.QueryRow(
		"SELECT student_id FROM projects WHERE id=?",
		input.ProjectID,
	).Scan(&ownerID)

	if err != nil || ownerID != me.ID {
		fail(w, 403, "Only the project owner can invite.")
		return
	}

	var exists int
	err = tx.QueryRow(`
        SELECT COUNT(*) FROM users
        WHERE id=? AND role='student' AND verified=1
    `, input.StudentID).Scan(&exists)

	if err != nil || exists != 1 {
		fail(w, 400, "Choose a verified student.")
		return
	}

	var duplicate int
	err = tx.QueryRow(`
        SELECT
          (SELECT COUNT(*) FROM team_members
           WHERE project_id=? AND user_id=?)
        +
          (SELECT COUNT(*) FROM workspace_records
           WHERE project_id=? AND assignee_id=?
             AND kind='invitation' AND status='pending')
    `, input.ProjectID, input.StudentID,
		input.ProjectID, input.StudentID).Scan(&duplicate)

	if err != nil {
		fail(w, 500, "Could not check invitation.")
		return
	}
	if duplicate > 0 {
		fail(w, 409, "Student is already invited or in your team.")
		return
	}

	var used int
	err = tx.QueryRow(`
        SELECT
          (SELECT COUNT(*) FROM team_members WHERE project_id=?)
        +
          (SELECT COUNT(*) FROM workspace_records
           WHERE project_id=? AND kind='invitation'
             AND status='pending')
    `, input.ProjectID, input.ProjectID).Scan(&used)

	if err != nil {
		fail(w, 500, "Could not check team size.")
		return
	}
	if used >= 2 {
		fail(w, 409, "Maximum 2 students. Pending invitations count.")
		return
	}

	_, err = tx.Exec(`
        INSERT INTO workspace_records
        (project_id, author_id, kind, title, body,
         status, assignee_id, updated_at)
        VALUES (?, ?, 'invitation', 'Team invitation', '',
                'pending', ?, ?)
    `, input.ProjectID, me.ID, input.StudentID, time.Now().Unix())

	if err != nil {
		fail(w, 500, "Could not save invitation.")
		return
	}

	if err = tx.Commit(); err != nil {
		fail(w, 500, "Could not finish invitation.")
		return
	}

	reply(w, 201, map[string]string{
		"message": "Invitation sent. Waiting for student acceptance.",
	})
}
