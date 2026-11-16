package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Adds the two-stage supervision workflow fields to existing SQLite databases.
func migrateSupervisionWorkflow(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(projects)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notNull, primary int
		var name, typ string
		var def sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notNull, &def, &primary); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if !existing["supervision_status"] {
		if _, err = db.Exec("ALTER TABLE projects ADD COLUMN supervision_status TEXT NOT NULL DEFAULT 'none'"); err != nil {
			return err
		}
	}
	if !existing["supervision_feedback"] {
		if _, err = db.Exec("ALTER TABLE projects ADD COLUMN supervision_feedback TEXT NOT NULL DEFAULT ''"); err != nil {
			return err
		}
	}
	return nil
}

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
		"SELECT status FROM projects WHERE id=? AND (student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=projects.id AND user_id=?)) AND supervision_status='accepted'",
		id,
		u.ID, u.ID,
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
		  AND (student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=projects.id AND user_id=?))
		  AND supervision_status='accepted'
		  AND status IN ('draft','changes_requested')
	`,
		in.ProblemStatement,
		in.Objectives,
		in.Methodology,
		in.ExpectedOutcomes,
		in.ScopeLimitations,
		time.Now().Unix(),
		id,
		u.ID, u.ID,
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
		  AND (student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=projects.id AND user_id=?))
		  AND supervision_status='accepted'
		  AND proposal_started=1
	`,
		id,
		u.ID, u.ID,
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
		  AND (student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=projects.id AND user_id=?))
		  AND supervision_status='accepted'
		  AND proposal_started=1
		  AND status IN ('draft','changes_requested')
		  AND teacher_id=?
	`,
		time.Now().Unix(),
		id,
		u.ID, u.ID,
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
	// Validate the profile of the project owner AND every confirmed team member.
	// A supervision request must never be sent if any member is incomplete.
	missingRows, err := a.db.Query(`
		SELECT u.name
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
		WHERE
			NOT EXISTS (
				SELECT 1
				FROM student_academic_records ar
				WHERE ar.user_id = u.id
			)
			OR NOT EXISTS (
				SELECT 1
				FROM student_skills ss
				WHERE ss.user_id = u.id
			)
		ORDER BY u.name
	`, id, id)
	if err != nil {
		fail(w, 500, "Could not validate team academic information.")
		return
	}
	missingMembers := []string{}
	for missingRows.Next() {
		var name string
		if err := missingRows.Scan(&name); err != nil {
			missingRows.Close()
			fail(w, 500, "Could not read team academic information.")
			return
		}
		missingMembers = append(missingMembers, name)
	}
	if err := missingRows.Err(); err != nil {
		missingRows.Close()
		fail(w, 500, "Could not read team academic information.")
		return
	}
	missingRows.Close()

	if len(missingMembers) > 0 {
		fail(
			w,
			400,
			"Supervision request not sent. These team members must complete Academic Profile with at least one year-wise CGPA record and at least one skill: "+strings.Join(missingMembers, ", ")+".",
		)
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

	var status, supervisionStatus string
	var existingTeacher int64

	err = a.db.QueryRow(`
		SELECT status, COALESCE(teacher_id, 0), COALESCE(supervision_status, 'none')
		FROM projects
		WHERE id=?
	`, id).Scan(&status, &existingTeacher, &supervisionStatus)

	if err != nil {
		fail(w, 404, "Project not found.")
		return
	}

	if status != "draft" && status != "changes_requested" {
		fail(w, 409, "The project is already submitted or approved.")
		return
	}

	if existingTeacher > 0 && existingTeacher != input.TeacherID && supervisionStatus != "declined" {
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
		SET teacher_id=?, supervision_status='pending', supervision_feedback='', updated_at=?
		WHERE id=?
		  AND status IN ('draft','changes_requested')
		  AND COALESCE(supervision_status,'none') IN ('none','declined')
	`, input.TeacherID, time.Now().Unix(), id)

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

	a.notifyProjectTeam(id, u.ID, "request", "New supervision request: "+strings.TrimSpace(projectTitle(a.db, id)))
	reply(w, 200, map[string]string{"message": "Supervision request sent. The full proposal will unlock after the supervisor accepts."})
}

func projectTitle(db *Database, id int64) string {
	var title string
	_ = db.QueryRow("SELECT title FROM projects WHERE id=?", id).Scan(&title)
	return title
}

func (a *App) reviewSupervisionRequest(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "teacher" {
		fail(w, 403, "Only teachers can decide supervision requests.")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "Invalid project.")
		return
	}
	var in struct {
		Status   string `json:"status"`
		Feedback string `json:"feedback"`
	}
	if !decode(w, r, &in) {
		return
	}
	in.Feedback = strings.TrimSpace(in.Feedback)
	if in.Status != "accepted" && in.Status != "declined" {
		fail(w, 400, "Choose accept or decline.")
		return
	}
	if in.Status == "declined" && len(in.Feedback) < 5 {
		fail(w, 400, "Please provide a short reason for declining.")
		return
	}
	result, err := a.db.Exec("UPDATE projects SET supervision_status=?, supervision_feedback=?, updated_at=? WHERE id=? AND teacher_id=? AND supervision_status='pending'", in.Status, in.Feedback, time.Now().Unix(), id, u.ID)
	if err != nil {
		fail(w, 500, "Could not save the supervision decision.")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		fail(w, 409, "This supervision request is no longer pending.")
		return
	}
	msg := "Supervision request accepted. Your team can now submit its full research proposal."
	if in.Status == "declined" {
		msg = "Supervision request declined. Feedback: " + in.Feedback
	}
	a.notifyProjectTeam(id, u.ID, "submission", msg)
	reply(w, 200, map[string]string{"message": "Decision saved and shared with the team."})
}

func (a *App) notifyProjectTeam(projectID, actorID int64, kind, message string) {
	_, _ = a.db.Exec(`INSERT INTO workspace_notifications(user_id,project_id,kind,message,created_at) SELECT DISTINCT member_id,?,?,?,? FROM (SELECT student_id AS member_id FROM projects WHERE id=? UNION SELECT user_id AS member_id FROM team_members WHERE project_id=?) members`, projectID, kind, message, time.Now().Unix(), projectID, projectID)
}
