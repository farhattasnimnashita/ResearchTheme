package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// migrateEthicsReviews is safe to run repeatedly. It creates the table for new
// installs and adds supervisor-review fields to older ethics_reviews tables.
func migrateEthicsReviews(db *Database) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ethics_reviews (
			project_id BIGINT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
			student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			ethics_required TEXT NOT NULL DEFAULT 'Not sure',
			committee TEXT NOT NULL DEFAULT '',
			application_reference TEXT NOT NULL DEFAULT '',
			review_status TEXT NOT NULL DEFAULT 'Not started',
			submission_date TEXT NOT NULL DEFAULT '',
			decision_date TEXT NOT NULL DEFAULT '',
			approval_number TEXT NOT NULL DEFAULT '',
			expiry_date TEXT NOT NULL DEFAULT '',
			participants_involved TEXT NOT NULL DEFAULT 'Not sure',
			consent_method TEXT NOT NULL DEFAULT '',
			consent_plan TEXT NOT NULL DEFAULT '',
			committee_conditions TEXT NOT NULL DEFAULT '',
			document_url TEXT NOT NULL DEFAULT '',
			supervisor_review_status TEXT NOT NULL DEFAULT 'Not requested',
			supervisor_comment TEXT NOT NULL DEFAULT '',
			supervisor_requested_at BIGINT NOT NULL DEFAULT 0,
			supervisor_reviewed_at BIGINT NOT NULL DEFAULT 0,
			updated_at BIGINT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// The first version of this feature may already have created the table.
	// Detect its columns portably, then add only the newer supervisor fields.
	rowsQuery := "PRAGMA table_info(ethics_reviews)"
	if db.postgres {
		rowsQuery = `SELECT column_name FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = 'ethics_reviews'`
	}

	rows, err := db.Query(rowsQuery)
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		if db.postgres {
			var name string
			if err := rows.Scan(&name); err != nil {
				rows.Close()
				return err
			}
			existing[name] = true
		} else {
			var cid, notNull, primary int
			var name, kind string
			var defaultValue any
			if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primary); err != nil {
				rows.Close()
				return err
			}
			existing[name] = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	columns := []struct {
		name       string
		definition string
	}{
		{"supervisor_review_status", "TEXT NOT NULL DEFAULT 'Not requested'"},
		{"supervisor_comment", "TEXT NOT NULL DEFAULT ''"},
		{"supervisor_requested_at", "BIGINT NOT NULL DEFAULT 0"},
		{"supervisor_reviewed_at", "BIGINT NOT NULL DEFAULT 0"},
	}
	for _, column := range columns {
		if existing[column.name] {
			continue
		}
		if _, err := db.Exec("ALTER TABLE ethics_reviews ADD COLUMN " + column.name + " " + column.definition); err != nil {
			return err
		}
	}
	return nil
}

type ethicsReviewRecord struct {
	ProjectID              int64  `json:"project_id"`
	ProjectTitle           string `json:"project_title"`
	StudentName            string `json:"student_name"`
	EthicsRequired         string `json:"ethics_required"`
	Committee              string `json:"committee"`
	ApplicationReference   string `json:"application_reference"`
	ReviewStatus           string `json:"review_status"`
	SubmissionDate         string `json:"submission_date"`
	DecisionDate           string `json:"decision_date"`
	ApprovalNumber         string `json:"approval_number"`
	ExpiryDate             string `json:"expiry_date"`
	ParticipantsInvolved   string `json:"participants_involved"`
	ConsentMethod          string `json:"consent_method"`
	ConsentPlan            string `json:"consent_plan"`
	CommitteeConditions    string `json:"committee_conditions"`
	DocumentURL            string `json:"document_url"`
	SupervisorReviewStatus string `json:"supervisor_review_status"`
	SupervisorComment      string `json:"supervisor_comment"`
	SupervisorRequestedAt  int64  `json:"supervisor_requested_at"`
	SupervisorReviewedAt   int64  `json:"supervisor_reviewed_at"`
	UpdatedAt              int64  `json:"updated_at"`
}

type ethicsReviewInput struct {
	Action                 string `json:"action"`
	ProjectID              int64  `json:"project_id"`
	EthicsRequired         string `json:"ethics_required"`
	Committee              string `json:"committee"`
	ApplicationReference   string `json:"application_reference"`
	ReviewStatus           string `json:"review_status"`
	SubmissionDate         string `json:"submission_date"`
	DecisionDate           string `json:"decision_date"`
	ApprovalNumber         string `json:"approval_number"`
	ExpiryDate             string `json:"expiry_date"`
	ParticipantsInvolved   string `json:"participants_involved"`
	ConsentMethod          string `json:"consent_method"`
	ConsentPlan            string `json:"consent_plan"`
	CommitteeConditions    string `json:"committee_conditions"`
	DocumentURL            string `json:"document_url"`
	SupervisorReviewStatus string `json:"supervisor_review_status"`
	SupervisorComment      string `json:"supervisor_comment"`
}

func ethicsChoice(value string, choices ...string) bool {
	for _, choice := range choices {
		if value == choice {
			return true
		}
	}
	return false
}

func validEthicsDate(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validEthicsURL(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}

func notifyEthicsUser(db *Database, userID, projectID int64, kind, message string) {
	if userID <= 0 || projectID <= 0 {
		return
	}
	_, err := db.Exec(`INSERT INTO workspace_notifications
		(user_id, project_id, kind, message, read_at, created_at)
		VALUES (?, ?, ?, ?, 0, ?)`, userID, projectID, kind, message, time.Now().Unix())
	if err != nil {
		// Saving the ethics record must not fail just because a notification could not be created.
		log.Printf("Could not create ethics notification: %v", err)
	}
}

// GET returns ethics records visible to the logged-in student/team or assigned supervisor.
// POST supports three actions: save (student), request_review (student), and review (assigned teacher).
func (a *App) researchEthics(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}

	if r.Method == http.MethodGet {
		var query string
		var args []any
		if u.Role == "teacher" {
			query = `SELECT er.project_id, p.title, s.name,
				er.ethics_required, er.committee, er.application_reference,
				er.review_status, er.submission_date, er.decision_date,
				er.approval_number, er.expiry_date, er.participants_involved,
				er.consent_method, er.consent_plan, er.committee_conditions,
				er.document_url, er.supervisor_review_status, er.supervisor_comment,
				er.supervisor_requested_at, er.supervisor_reviewed_at, er.updated_at
				FROM ethics_reviews er
				JOIN projects p ON p.id = er.project_id
				JOIN users s ON s.id = p.student_id
				WHERE p.teacher_id = ?
				ORDER BY CASE WHEN er.supervisor_review_status = 'Pending' THEN 0 ELSE 1 END,
					er.updated_at DESC, p.title`
			args = []any{u.ID}
		} else if u.Role == "student" {
			query = `SELECT er.project_id, p.title, s.name,
				er.ethics_required, er.committee, er.application_reference,
				er.review_status, er.submission_date, er.decision_date,
				er.approval_number, er.expiry_date, er.participants_involved,
				er.consent_method, er.consent_plan, er.committee_conditions,
				er.document_url, er.supervisor_review_status, er.supervisor_comment,
				er.supervisor_requested_at, er.supervisor_reviewed_at, er.updated_at
				FROM ethics_reviews er
				JOIN projects p ON p.id = er.project_id
				JOIN users s ON s.id = p.student_id
				WHERE p.student_id = ? OR EXISTS (
					SELECT 1 FROM team_members tm
					WHERE tm.project_id = p.id AND tm.user_id = ?
				)
				ORDER BY er.updated_at DESC, p.title`
			args = []any{u.ID, u.ID}
		} else {
			fail(w, 403, "This role cannot access research ethics records.")
			return
		}

		rows, err := a.db.Query(query, args...)
		if err != nil {
			fail(w, 500, "Could not load ethics records.")
			return
		}
		defer rows.Close()

		records := make([]ethicsReviewRecord, 0)
		for rows.Next() {
			var record ethicsReviewRecord
			if err := rows.Scan(
				&record.ProjectID, &record.ProjectTitle, &record.StudentName,
				&record.EthicsRequired, &record.Committee, &record.ApplicationReference,
				&record.ReviewStatus, &record.SubmissionDate, &record.DecisionDate,
				&record.ApprovalNumber, &record.ExpiryDate, &record.ParticipantsInvolved,
				&record.ConsentMethod, &record.ConsentPlan, &record.CommitteeConditions,
				&record.DocumentURL, &record.SupervisorReviewStatus, &record.SupervisorComment,
				&record.SupervisorRequestedAt, &record.SupervisorReviewedAt, &record.UpdatedAt,
			); err != nil {
				fail(w, 500, "Could not read ethics records.")
				return
			}
			records = append(records, record)
		}
		if err := rows.Err(); err != nil {
			fail(w, 500, "Could not read ethics records.")
			return
		}
		reply(w, 200, records)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		fail(w, 405, "Method not allowed.")
		return
	}

	var in ethicsReviewInput
	if !decode(w, r, &in) {
		return
	}
	in.Action = strings.TrimSpace(in.Action)

	// Only the assigned supervisor may write supervisor feedback.
	if in.Action == "review" {
		if u.Role != "teacher" {
			fail(w, 403, "Only the assigned supervisor can submit a supervisor review.")
			return
		}
		in.SupervisorReviewStatus = strings.TrimSpace(in.SupervisorReviewStatus)
		in.SupervisorComment = strings.TrimSpace(in.SupervisorComment)
		if in.ProjectID <= 0 || !ethicsChoice(in.SupervisorReviewStatus, "Changes requested", "Reviewed") {
			fail(w, 400, "Choose Changes requested or Reviewed.")
			return
		}
		if len(in.SupervisorComment) > 5000 {
			fail(w, 400, "Supervisor comments cannot exceed 5000 characters.")
			return
		}
		if in.SupervisorReviewStatus == "Changes requested" && in.SupervisorComment == "" {
			fail(w, 400, "Add feedback explaining which changes are needed.")
			return
		}

		var ownerID int64
		var projectTitle string
		err := a.db.QueryRow(`SELECT p.student_id, p.title
			FROM projects p JOIN ethics_reviews er ON er.project_id = p.id
			WHERE p.id = ? AND p.teacher_id = ?`, in.ProjectID, u.ID).Scan(&ownerID, &projectTitle)
		if err != nil {
			fail(w, 404, "Ethics record not found for a project assigned to you.")
			return
		}

		now := time.Now().Unix()
		result, err := a.db.Exec(`UPDATE ethics_reviews
			SET supervisor_review_status = ?, supervisor_comment = ?,
				supervisor_reviewed_at = ?, updated_at = ?
			WHERE project_id = ?`, in.SupervisorReviewStatus, in.SupervisorComment, now, now, in.ProjectID)
		if err != nil {
			fail(w, 500, "Could not save supervisor feedback.")
			return
		}
		if changed, _ := result.RowsAffected(); changed == 0 {
			fail(w, 404, "Ethics record not found.")
			return
		}
		notifyEthicsUser(a.db, ownerID, in.ProjectID, "ethics_review", "Your supervisor updated the ethics review for: "+projectTitle)
		reply(w, 200, map[string]string{"message": "Supervisor review saved. This is not official ethics approval."})
		return
	}

	if u.Role != "student" {
		fail(w, 403, "Only students can update ethics applications or request a supervisor review.")
		return
	}
	if in.ProjectID <= 0 {
		fail(w, 400, "Please select a research project.")
		return
	}

	// Verify this student belongs to the project, and learn its official owner/supervisor.
	var ownerID, teacherID int64
	var projectTitle string
	err = a.db.QueryRow(`SELECT p.student_id, COALESCE(p.teacher_id, 0), p.title
		FROM projects p
		WHERE p.id = ? AND (p.student_id = ? OR EXISTS (
			SELECT 1 FROM team_members tm WHERE tm.project_id = p.id AND tm.user_id = ?
		))`, in.ProjectID, u.ID, u.ID).Scan(&ownerID, &teacherID, &projectTitle)
	if err != nil {
		fail(w, 404, "Project not found or you do not have access to it.")
		return
	}

	if in.Action == "request_review" {
		if teacherID <= 0 {
			fail(w, 409, "Choose an assigned supervisor for this project before requesting a review.")
			return
		}
		var existing int
		err := a.db.QueryRow("SELECT 1 FROM ethics_reviews WHERE project_id = ?", in.ProjectID).Scan(&existing)
		if err == sql.ErrNoRows {
			fail(w, 409, "Save the ethics record before requesting supervisor review.")
			return
		}
		if err != nil {
			fail(w, 500, "Could not check the ethics record.")
			return
		}
		var currentStatus string
		if err := a.db.QueryRow("SELECT supervisor_review_status FROM ethics_reviews WHERE project_id = ?", in.ProjectID).Scan(&currentStatus); err != nil {
			fail(w, 500, "Could not check the current review status.")
			return
		}
		if currentStatus == "Pending" {
			fail(w, 409, "A supervisor review request is already pending.")
			return
		}
		now := time.Now().Unix()
		if _, err := a.db.Exec(`UPDATE ethics_reviews
			SET supervisor_review_status = 'Pending', supervisor_comment = '',
				supervisor_requested_at = ?, supervisor_reviewed_at = 0, updated_at = ?
			WHERE project_id = ?`, now, now, in.ProjectID); err != nil {
			fail(w, 500, "Could not request supervisor review.")
			return
		}
		notifyEthicsUser(a.db, teacherID, in.ProjectID, "ethics_review", "Supervisor review requested for: "+projectTitle)
		reply(w, 200, map[string]string{"message": "Supervisor review requested. The teacher's review is separate from official ethics approval."})
		return
	}

	// Empty action is accepted as "save" for compatibility with simple forms.
	if in.Action != "" && in.Action != "save" {
		fail(w, 400, "Unknown ethics action.")
		return
	}

	in.EthicsRequired = strings.TrimSpace(in.EthicsRequired)
	in.Committee = strings.TrimSpace(in.Committee)
	in.ApplicationReference = strings.TrimSpace(in.ApplicationReference)
	in.ReviewStatus = strings.TrimSpace(in.ReviewStatus)
	in.SubmissionDate = strings.TrimSpace(in.SubmissionDate)
	in.DecisionDate = strings.TrimSpace(in.DecisionDate)
	in.ApprovalNumber = strings.TrimSpace(in.ApprovalNumber)
	in.ExpiryDate = strings.TrimSpace(in.ExpiryDate)
	in.ParticipantsInvolved = strings.TrimSpace(in.ParticipantsInvolved)
	in.ConsentMethod = strings.TrimSpace(in.ConsentMethod)
	in.ConsentPlan = strings.TrimSpace(in.ConsentPlan)
	in.CommitteeConditions = strings.TrimSpace(in.CommitteeConditions)
	in.DocumentURL = strings.TrimSpace(in.DocumentURL)

	if !ethicsChoice(in.EthicsRequired, "Yes", "No", "Not sure") {
		fail(w, 400, "Select whether ethics review is required.")
		return
	}
	if !ethicsChoice(in.ReviewStatus, "Not started", "Preparing", "Submitted", "Changes requested", "Approved", "Rejected", "Exempt", "Not required") {
		fail(w, 400, "Select a valid ethics committee status.")
		return
	}
	if !ethicsChoice(in.ParticipantsInvolved, "Yes", "No", "Not sure") {
		fail(w, 400, "Select whether human participants are involved.")
		return
	}
	if len(in.Committee) > 250 || len(in.ApplicationReference) > 250 || len(in.ApprovalNumber) > 250 || len(in.ConsentMethod) > 250 || len(in.ConsentPlan) > 5000 || len(in.CommitteeConditions) > 5000 || len(in.DocumentURL) > 2000 {
		fail(w, 400, "One or more fields are too long.")
		return
	}
	if !validEthicsDate(in.SubmissionDate) || !validEthicsDate(in.DecisionDate) || !validEthicsDate(in.ExpiryDate) {
		fail(w, 400, "Please enter valid dates.")
		return
	}
	if in.SubmissionDate != "" && in.DecisionDate != "" && in.DecisionDate < in.SubmissionDate {
		fail(w, 400, "Decision date cannot be before submission date.")
		return
	}
	if in.DecisionDate != "" && in.ExpiryDate != "" && in.ExpiryDate < in.DecisionDate {
		fail(w, 400, "Expiry date cannot be before decision date.")
		return
	}
	if !validEthicsURL(in.DocumentURL) {
		fail(w, 400, "Document link must be a valid HTTP or HTTPS URL.")
		return
	}
	if in.EthicsRequired == "No" && in.ReviewStatus != "Not required" {
		fail(w, 400, "When ethics review is not required, set status to Not required.")
		return
	}
	if in.ReviewStatus == "Not required" && in.EthicsRequired != "No" {
		fail(w, 400, "Choose Ethics Review Required = No before setting Not required.")
		return
	}
	if ethicsChoice(in.ReviewStatus, "Submitted", "Changes requested", "Approved", "Rejected") && in.SubmissionDate == "" {
		fail(w, 400, "Enter the application submission date for this status.")
		return
	}
	if in.ReviewStatus == "Approved" && (in.ApprovalNumber == "" || in.DecisionDate == "" || in.DocumentURL == "") {
		fail(w, 400, "Approved status requires the official approval number, decision date and document link.")
		return
	}
	if in.ReviewStatus == "Exempt" && (in.DecisionDate == "" || in.DocumentURL == "") {
		fail(w, 400, "Exempt status requires the official decision date and exemption evidence link.")
		return
	}

	now := time.Now().Unix()
	_, err = a.db.Exec(`INSERT INTO ethics_reviews (
		project_id, student_id, ethics_required, committee, application_reference,
		review_status, submission_date, decision_date, approval_number, expiry_date,
		participants_involved, consent_method, consent_plan, committee_conditions,
		document_url, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(project_id) DO UPDATE SET
		student_id = excluded.student_id,
		ethics_required = excluded.ethics_required,
		committee = excluded.committee,
		application_reference = excluded.application_reference,
		review_status = excluded.review_status,
		submission_date = excluded.submission_date,
		decision_date = excluded.decision_date,
		approval_number = excluded.approval_number,
		expiry_date = excluded.expiry_date,
		participants_involved = excluded.participants_involved,
		consent_method = excluded.consent_method,
		consent_plan = excluded.consent_plan,
		committee_conditions = excluded.committee_conditions,
		document_url = excluded.document_url,
		supervisor_review_status = 'Not requested',
		supervisor_comment = '',
		supervisor_requested_at = 0,
		supervisor_reviewed_at = 0,
		updated_at = excluded.updated_at`,
		in.ProjectID, ownerID, in.EthicsRequired, in.Committee, in.ApplicationReference,
		in.ReviewStatus, in.SubmissionDate, in.DecisionDate, in.ApprovalNumber, in.ExpiryDate,
		in.ParticipantsInvolved, in.ConsentMethod, in.ConsentPlan, in.CommitteeConditions,
		in.DocumentURL, now)
	if err != nil {
		fail(w, 500, "Could not save the ethics record.")
		return
	}
	reply(w, 200, map[string]any{"message": "Ethics record saved successfully.", "project_id": in.ProjectID})
}
