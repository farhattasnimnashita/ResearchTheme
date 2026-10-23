package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Existing SQLite database-এ Proposal-এর নতুন column যোগ করে।
func migrateProjectProposalColumns(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(projects)")
	if err != nil {
		return err
	}

	existing := make(map[string]bool)

	for rows.Next() {
		var cid, notNull, primary int
		var name, dataType string
		var defaultValue sql.NullString

		if err := rows.Scan(
			&cid,
			&name,
			&dataType,
			&notNull,
			&defaultValue,
			&primary,
		); err != nil {
			rows.Close()
			return err
		}

		existing[name] = true
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}

	if err := rows.Close(); err != nil {
		return err
	}

	columns := []struct {
		name string
		sql  string
	}{
		{
			"problem_statement",
			"ALTER TABLE projects ADD COLUMN problem_statement TEXT NOT NULL DEFAULT ''",
		},
		{
			"objectives",
			"ALTER TABLE projects ADD COLUMN objectives TEXT NOT NULL DEFAULT ''",
		},
		{
			"methodology",
			"ALTER TABLE projects ADD COLUMN methodology TEXT NOT NULL DEFAULT ''",
		},
		{
			"expected_outcomes",
			"ALTER TABLE projects ADD COLUMN expected_outcomes TEXT NOT NULL DEFAULT ''",
		},
		{
			"scope_limitations",
			"ALTER TABLE projects ADD COLUMN scope_limitations TEXT NOT NULL DEFAULT ''",
		},
		{
			"proposal_started",
			"ALTER TABLE projects ADD COLUMN proposal_started INTEGER NOT NULL DEFAULT 0",
		},
	}

	for _, column := range columns {
		if existing[column.name] {
			continue
		}

		if _, err := db.Exec(column.sql); err != nil {
			return err
		}
	}

	return nil
}

// Proposal draft-এর text size যাচাই করে।
func proposalLengthOK(value string, min, max int) bool {
	n := len(strings.TrimSpace(value))
	return n >= min && n <= max
}

// Student-এর Proposal draft database-এ save করে।
func (a *App) saveProjectProposal(
	w http.ResponseWriter,
	r *http.Request,
) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	if u.Role != "student" {
		fail(w, 403, "Only students can edit a research proposal.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "Invalid project.")
		return
	}

	// Project-টি এই student-এর কি না এবং edit করা যাবে কি না।
	var status string

	err = a.db.QueryRow(
		"SELECT status FROM projects WHERE id=? AND student_id=?",
		id,
		u.ID,
	).Scan(&status)

	if err != nil {
		fail(w, 404, "Project not found.")
		return
	}

	if status != "draft" && status != "changes_requested" {
		fail(w, 409, "This proposal is no longer editable.")
		return
	}

	var in struct {
		TeacherID        int64  `json:"teacher_id"`
		ProblemStatement string `json:"problem_statement"`
		Objectives       string `json:"objectives"`
		Methodology      string `json:"methodology"`
		ExpectedOutcomes string `json:"expected_outcomes"`
		ScopeLimitations string `json:"scope_limitations"`
	}

	if !decode(w, r, &in) {
		return
	}

	in.ProblemStatement = strings.TrimSpace(in.ProblemStatement)
	in.Objectives = strings.TrimSpace(in.Objectives)
	in.Methodology = strings.TrimSpace(in.Methodology)
	in.ExpectedOutcomes = strings.TrimSpace(in.ExpectedOutcomes)
	in.ScopeLimitations = strings.TrimSpace(in.ScopeLimitations)

	if in.TeacherID < 0 ||
		len(in.ProblemStatement) > 6000 ||
		len(in.Objectives) > 4000 ||
		len(in.Methodology) > 6000 ||
		len(in.ExpectedOutcomes) > 4000 ||
		len(in.ScopeLimitations) > 3000 {
		fail(w, 400, "Some proposal fields are too long or invalid.")
		return
	}

	// Supervisor selection is managed by selectProjectSupervisor.
	// Saving proposal text must preserve the existing assignment.
	result, err := a.db.Exec(`
		UPDATE projects
		SET problem_statement=?,
		    objectives=?,
		    methodology=?,
		    expected_outcomes=?,
		    scope_limitations=?,
		    proposal_started=1,
		    updated_at=?
		WHERE id=?
		  AND student_id=?
		  AND status IN ('draft','changes_requested')
	`,
		in.ProblemStatement,
		in.Objectives,
		in.Methodology,
		in.ExpectedOutcomes,
		in.ScopeLimitations,
		time.Now().Unix(),
		id,
		u.ID,
	)

	if err != nil {
		fail(w, 500, "Could not save the proposal draft.")
		return
	}

	n, err := result.RowsAffected()
	if err != nil || n == 0 {
		fail(w, 409, "This project is no longer editable.")
		return
	}

	reply(w, 200, map[string]any{
		"id":      id,
		"message": "Proposal draft saved successfully.",
	})
}

// পূর্ণ Proposal supervisor-এর review-এর জন্য submit করে।
func (a *App) submitDetailedProjectProposal(
	w http.ResponseWriter,
	r *http.Request,
) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	if u.Role != "student" {
		fail(w, 403, "Only students can submit a proposal.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "Invalid project.")
		return
	}

	var teacherID int64
	var problem, objectives, method, outcomes, scope, status string

	err = a.db.QueryRow(`
		SELECT COALESCE(teacher_id,0),
		       problem_statement,
		       objectives,
		       methodology,
		       expected_outcomes,
		       scope_limitations,
		       status
		FROM projects
		WHERE id=?
		  AND student_id=?
		  AND proposal_started=1
	`,
		id,
		u.ID,
	).Scan(
		&teacherID,
		&problem,
		&objectives,
		&method,
		&outcomes,
		&scope,
		&status,
	)

	if err != nil {
		fail(w, 404, "Proposal draft not found. Save your draft first.")
		return
	}

	if status != "draft" && status != "changes_requested" {
		fail(w, 409, "This proposal cannot be submitted in its current status.")
		return
	}

	// Submit করার সময় প্রয়োজনীয় তথ্য থাকতে হবে।
	if !proposalLengthOK(problem, 20, 6000) {
		fail(w, 400, "Problem Statement must contain 20–6000 characters.")
		return
	}

	if !proposalLengthOK(objectives, 10, 4000) {
		fail(w, 400, "Objectives must contain 10–4000 characters.")
		return
	}

	if !proposalLengthOK(method, 20, 6000) {
		fail(w, 400, "Methodology must contain 20–6000 characters.")
		return
	}

	if !proposalLengthOK(outcomes, 10, 4000) {
		fail(w, 400, "Expected Outcomes must contain 10–4000 characters.")
		return
	}

	if !proposalLengthOK(scope, 10, 3000) {
		fail(w, 400, "Scope and Limitations must contain 10–3000 characters.")
		return
	}

	if teacherID <= 0 {
		fail(w, 400, "Choose a supervisor before submitting.")
		return
	}

	var verifiedTeacher int64

	err = a.db.QueryRow(
		"SELECT id FROM users WHERE id=? AND role='teacher' AND verified=1",
		teacherID,
	).Scan(&verifiedTeacher)

	if err != nil {
		fail(w, 400, "The selected supervisor is not available.")
		return
	}

	result, err := a.db.Exec(`
		UPDATE projects
		SET status='submitted', updated_at=?
		WHERE id=?
		  AND student_id=?
		  AND proposal_started=1
		  AND status IN ('draft','changes_requested')
		  AND teacher_id=?
	`,
		time.Now().Unix(),
		id,
		u.ID,
		teacherID,
	)

	if err != nil {
		fail(w, 500, "Could not submit the proposal.")
		return
	}

	n, err := result.RowsAffected()
	if err != nil || n == 0 {
		fail(w, 409, "The proposal has changed. Refresh and try again.")
		return
	}

	// Existing notification workflow-তেই supervisor-কে জানানো হবে।
	a.notifyProject(
		id,
		u.ID,
		"request",
		"A research proposal is ready for your review.",
	)

	reply(w, 200, map[string]any{
		"message": "Proposal submitted successfully.",
		"status":  "submitted",
	})
}
func (a *App) selectProjectSupervisor(
	w http.ResponseWriter,
	r *http.Request,
) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	if u.Role != "student" {
		fail(w, 403, "Only students can select a supervisor.")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "Invalid project.")
		return
	}

	if !a.projectAccess(u, id) {
		fail(w, 403, "This project is not assigned to you.")
		return
	}

	var input struct {
		TeacherID int64 `json:"teacher_id"`
	}

	if !decode(w, r, &input) {
		return
	}

	if input.TeacherID <= 0 {
		fail(w, 400, "Choose a verified supervisor.")
		return
	}

	var status string
	var existingTeacher int64

	err = a.db.QueryRow(`
		SELECT status, COALESCE(teacher_id, 0)
		FROM projects
		WHERE id=?
	`, id).Scan(&status, &existingTeacher)

	if err != nil {
		fail(w, 404, "Project not found.")
		return
	}

	if status != "draft" && status != "changes_requested" {
		fail(w, 409, "The project is already submitted or approved.")
		return
	}

	if existingTeacher > 0 && existingTeacher != input.TeacherID {
		fail(
			w,
			409,
			"A supervisor is already selected for this team. All members must use the same supervisor.",
		)
		return
	}

	var verifiedTeacher int64
	err = a.db.QueryRow(`
		SELECT id
		FROM users
		WHERE id=? AND role='teacher' AND verified=1
	`, input.TeacherID).Scan(&verifiedTeacher)

	if err != nil {
		fail(w, 400, "Choose a verified supervisor.")
		return
	}

	result, err := a.db.Exec(`
		UPDATE projects
		SET teacher_id=?, updated_at=?
		WHERE id=?
		  AND status IN ('draft','changes_requested')
		  AND (teacher_id IS NULL OR teacher_id=?)
	`,
		input.TeacherID,
		time.Now().Unix(),
		id,
		input.TeacherID,
	)

	if err != nil {
		fail(w, 500, "Could not save the supervisor.")
		return
	}

	n, err := result.RowsAffected()
	if err != nil || n == 0 {
		fail(
			w,
			409,
			"The project or supervisor selection has changed. Refresh and try again.",
		)
		return
	}

	reply(w, 200, map[string]string{
		"message": "Supervisor fixed for this project and shared by all team members.",
	})
}
