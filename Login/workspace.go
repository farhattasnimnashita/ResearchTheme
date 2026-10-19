package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func migrateWorkspace(db *sql.DB) error {
	_, err := db.Exec(`
 CREATE TABLE IF NOT EXISTS password_resets(hash TEXT PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,expires INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS reset_requests(email_hash TEXT PRIMARY KEY,requested_at INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS researcher_profiles(user_id INTEGER PRIMARY KEY REFERENCES users(id),bio TEXT NOT NULL DEFAULT '',interests TEXT NOT NULL DEFAULT '',expertise TEXT NOT NULL DEFAULT '',notify_requests INTEGER NOT NULL DEFAULT 1,notify_submissions INTEGER NOT NULL DEFAULT 1,notify_meetings INTEGER NOT NULL DEFAULT 1);
 CREATE TABLE IF NOT EXISTS team_members(project_id INTEGER NOT NULL REFERENCES projects(id),user_id INTEGER NOT NULL REFERENCES users(id),PRIMARY KEY(project_id,user_id));
 CREATE TABLE IF NOT EXISTS project_permissions(project_id INTEGER PRIMARY KEY REFERENCES projects(id),teacher_id INTEGER REFERENCES users(id),can_evaluate INTEGER NOT NULL DEFAULT 0,can_extend INTEGER NOT NULL DEFAULT 0);
 CREATE TABLE IF NOT EXISTS workspace_records(id INTEGER PRIMARY KEY,project_id INTEGER REFERENCES projects(id),author_id INTEGER NOT NULL REFERENCES users(id),kind TEXT NOT NULL,title TEXT NOT NULL,body TEXT NOT NULL DEFAULT '',link TEXT NOT NULL DEFAULT '',status TEXT NOT NULL,feedback TEXT NOT NULL DEFAULT '',due_at INTEGER NOT NULL DEFAULT 0,progress INTEGER NOT NULL DEFAULT 0,score INTEGER NOT NULL DEFAULT 0,criteria TEXT NOT NULL DEFAULT '',assignee_id INTEGER REFERENCES users(id),parent_id INTEGER REFERENCES workspace_records(id),updated_at INTEGER NOT NULL);
 CREATE INDEX IF NOT EXISTS workspace_project ON workspace_records(project_id,kind,updated_at);
 CREATE TABLE IF NOT EXISTS workspace_history(id INTEGER PRIMARY KEY,record_id INTEGER NOT NULL REFERENCES workspace_records(id),actor_id INTEGER NOT NULL REFERENCES users(id),snapshot TEXT NOT NULL,created_at INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS workspace_files(id INTEGER PRIMARY KEY,project_id INTEGER NOT NULL REFERENCES projects(id),author_id INTEGER NOT NULL REFERENCES users(id),name TEXT NOT NULL,media_type TEXT NOT NULL,data BLOB NOT NULL,created_at INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS workspace_notifications(id INTEGER PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id),project_id INTEGER NOT NULL REFERENCES projects(id),kind TEXT NOT NULL,message TEXT NOT NULL,read_at INTEGER NOT NULL DEFAULT 0,created_at INTEGER NOT NULL);
 `)
	return err
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

type ResearcherProfile struct {
	Bio               string `json:"bio"`
	Interests         string `json:"interests"`
	Expertise         string `json:"expertise"`
	NotifyRequests    bool   `json:"notify_requests"`
	NotifySubmissions bool   `json:"notify_submissions"`
	NotifyMeetings    bool   `json:"notify_meetings"`
	Photo             string `json:"photo"`
	Qualification     string `json:"qualification"`
}

func (a *App) researcherProfile(w http.ResponseWriter, r *http.Request) {
	u, err := a.user(r)
	if err != nil {
		fail(w, 401, "Please log in.")
		return
	}
	p := ResearcherProfile{NotifyRequests: true, NotifySubmissions: true, NotifyMeetings: true}
	if r.Method == "POST" {
		if !decode(w, r, &p) {
			return
		}
		p.Bio = strings.TrimSpace(p.Bio)
		p.Interests = strings.TrimSpace(p.Interests)
		p.Expertise = strings.TrimSpace(p.Expertise)
		if len(p.Bio) > 3000 || len(p.Interests) > 1500 || len(p.Expertise) > 1500 || len(p.Qualification) > 2000 || len(p.Photo) > 450000 || (p.Photo != "" && !strings.HasPrefix(p.Photo, "data:image/")) {
			fail(w, 400, "Profile text is too long.")
			return
		}
		_, err = a.db.Exec(`INSERT INTO researcher_profiles(user_id,bio,interests,expertise,notify_requests,notify_submissions,notify_meetings,photo,qualification) VALUES(?,?,?,?,?,?,?,?,?) ON CONFLICT(user_id) DO UPDATE SET bio=excluded.bio,interests=excluded.interests,expertise=excluded.expertise,notify_requests=excluded.notify_requests,notify_submissions=excluded.notify_submissions,notify_meetings=excluded.notify_meetings,photo=excluded.photo,qualification=excluded.qualification`, u.ID, p.Bio, p.Interests, p.Expertise, boolInt(p.NotifyRequests), boolInt(p.NotifySubmissions), boolInt(p.NotifyMeetings), p.Photo, p.Qualification)
	} else {
		var req, sub, meet int64
		err = a.db.QueryRow("SELECT bio,interests,expertise,notify_requests,notify_submissions,notify_meetings,photo,qualification FROM researcher_profiles WHERE user_id=?", u.ID).Scan(&p.Bio, &p.Interests, &p.Expertise, &req, &sub, &meet, &p.Photo, &p.Qualification)
		p.NotifyRequests = req != 0
		p.NotifySubmissions = sub != 0
		p.NotifyMeetings = meet != 0
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
	}
	if err != nil {
		fail(w, 500, "Could not save or load profile.")
		return
	}
	reply(w, 200, p)
}
func (a *App) projectAccess(u User, id int64) bool {
	var n int
	if u.Role == "teacher" {
		a.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id=? AND teacher_id=? AND status!='draft'", id, u.ID).Scan(&n)
	} else {
		a.db.QueryRow("SELECT COUNT(*) FROM projects WHERE id=? AND (student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=projects.id AND user_id=?))", id, u.ID, u.ID).Scan(&n)
	}
	return n == 1
}
func projectID(r *http.Request) int64 {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	return id
}
func (a *App) workspaceMembers(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	id := projectID(r)
	if !a.projectAccess(u, id) {
		fail(w, 403, "This project is not assigned to you.")
		return
	}
	if r.Method == "POST" {
		var in struct {
			Email  string `json:"email"`
			Action string `json:"action"`
		}
		if !decode(w, r, &in) {
			return
		}
		// Membership is opt-in: a student can join only using their own email and a current member's invitation record.
		em, ok := email(in.Email)
		if !ok {
			fail(w, 400, "Enter a student email.")
			return
		}
		var owner int64
		a.db.QueryRow("SELECT student_id FROM projects WHERE id=?", id).Scan(&owner)
		if u.ID != owner || u.Role != "student" || in.Action != "invite" {
			fail(w, 403, "Only the project owner may invite members.")
			return
		}
		var student int64
		if a.db.QueryRow("SELECT id FROM users WHERE email=? AND role='student' AND verified=1", em).Scan(&student) != nil || student == owner {
			fail(w, 400, "Choose another registered, verified student.")
			return
		}
		_, e = a.db.Exec("INSERT INTO workspace_records(project_id,author_id,kind,title,body,status,assignee_id,updated_at) VALUES(?,?,'invitation','Team invitation','','pending',?,?)", id, u.ID, student, time.Now().Unix())
		if e != nil {
			fail(w, 500, "Could not save invitation.")
			return
		}
		reply(w, 201, map[string]string{"message": "Invitation saved. The student can accept from Research workspace → Invitations."})
		return
	}
	rows, e := a.db.Query("SELECT id,name,email FROM users WHERE id=(SELECT student_id FROM projects WHERE id=?) OR id IN(SELECT user_id FROM team_members WHERE project_id=?) ORDER BY name", id, id)
	if e != nil {
		fail(w, 500, "Could not load members.")
		return
	}
	defer rows.Close()
	members := []map[string]any{}
	for rows.Next() {
		var uid int64
		var name, em string
		if rows.Scan(&uid, &name, &em) != nil {
			fail(w, 500, "Could not read members.")
			return
		}
		members = append(members, map[string]any{"id": uid, "name": name, "email": em})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read members.")
		return
	}
	reply(w, 200, members)
}
func (a *App) invitations(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Student invitations only.")
		return
	}
	if r.Method == "POST" {
		var in struct {
			ID     int64 `json:"id"`
			Accept bool  `json:"accept"`
		}
		if !decode(w, r, &in) {
			return
		}
		tx, e := a.db.Begin()
		if e != nil {
			fail(w, 500, "Could not process invitation.")
			return
		}
		defer tx.Rollback()

		var projectID, ownerID int64
		var projectTitle string

		err := tx.QueryRow(`
    SELECT r.project_id, p.student_id, p.title
    FROM workspace_records r
    JOIN projects p ON p.id = r.project_id
    WHERE r.id = ?
      AND r.assignee_id = ?
      AND r.kind = 'invitation'
      AND r.status = 'pending'
`, in.ID, u.ID).Scan(
			&projectID,
			&ownerID,
			&projectTitle,
		)

		if err != nil {
			fail(w, 404, "Invitation not found or already processed.")
			return
		}

		status := "declined"
		message := u.Name + " rejected your invitation for " +
			projectTitle + ". Find another team member."

		if in.Accept {
			var memberCount int

			err = tx.QueryRow(`
        SELECT COUNT(*)
        FROM team_members
        WHERE project_id = ?
    `, projectID).Scan(&memberCount)

			if err != nil {
				fail(w, 500, "Could not check team size.")
				return
			}

			if memberCount >= 2 {
				fail(w, 409, "This team is already full.")
				return
			}

			_, err = tx.Exec(`
        INSERT INTO team_members
        (project_id, user_id)
        VALUES (?, ?) ON CONFLICT(project_id,user_id) DO NOTHING
    `, projectID, u.ID)

			if err != nil {
				fail(w, 500, "Could not join team.")
				return
			}

			status = "accepted"
			message = u.Name + " accepted your invitation for " +
				projectTitle + " and joined your team."
		}

		// Save the invitation decision.
		_, err = tx.Exec(`
    UPDATE workspace_records
    SET status = ?, updated_at = ?
    WHERE id = ? AND status = 'pending'
`, status, time.Now().Unix(), in.ID)

		if err != nil {
			fail(w, 500, "Could not update invitation.")
			return
		}

		// Notify the team owner about acceptance or rejection.
		_, err = tx.Exec(`
    INSERT INTO workspace_notifications
    (user_id, project_id, kind, message, created_at)
    VALUES (?, ?, 'team', ?, ?)
`, ownerID, projectID, message, time.Now().Unix())

		if err != nil {
			fail(w, 500, "Could not create notification.")
			return
		}

		if err = tx.Commit(); err != nil {
			fail(w, 500, "Could not finish invitation.")
			return
		}

		if in.Accept {
			reply(w, 200, map[string]string{
				"message": "Invitation accepted. You are now a team member.",
			})
		} else {
			reply(w, 200, map[string]string{
				"message": "Invitation rejected.",
			})
		}
		return

	}
	rows, e := a.db.Query("SELECT r.id,p.title,u.name FROM workspace_records r JOIN projects p ON p.id=r.project_id JOIN users u ON u.id=r.author_id WHERE r.kind='invitation' AND r.status='pending' AND r.assignee_id=?", u.ID)
	if e != nil {
		fail(w, 500, "Could not load invitations.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var title, name string
		if rows.Scan(&id, &title, &name) != nil {
			fail(w, 500, "Could not read invitations.")
			return
		}
		out = append(out, map[string]any{"id": id, "title": title, "name": name})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read invitations.")
		return
	}
	reply(w, 200, out)
}

type WorkspaceRecord struct {
	ID         int64  `json:"id"`
	ProjectID  int64  `json:"project_id"`
	AuthorID   int64  `json:"author_id"`
	Author     string `json:"author"`
	Kind       string `json:"kind"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	Link       string `json:"link"`
	Status     string `json:"status"`
	Feedback   string `json:"feedback"`
	DueAt      int64  `json:"due_at"`
	Progress   int    `json:"progress"`
	Score      int    `json:"score"`
	Criteria   string `json:"criteria"`
	AssigneeID int64  `json:"assignee_id"`
	ParentID   int64  `json:"parent_id"`
	UpdatedAt  int64  `json:"updated_at"`
}

var projectKinds = map[string]bool{"milestone": true, "task": true, "note": true, "submission": true, "feedback": true, "extension": true, "contribution": true, "paper": true, "experiment": true, "meeting": true, "blocker": true, "final_submission": true, "evaluation": true, "correction": true, "mentorship": true, "publication": true, "showcase": true}
var globalKinds = map[string]bool{"opportunity": true, "lab": true, "campus_problem": true}
var studentKinds = map[string]bool{"submission": true, "extension": true, "contribution": true, "paper": true, "experiment": true, "meeting": true, "blocker": true, "final_submission": true, "correction": true, "publication": true, "note": true}
var workspaceStatuses = map[string]bool{"pending": true, "planned": true, "active": true, "approved": true, "changes_requested": true, "rejected": true, "completed": true, "scheduled": true, "cancelled": true, "published": true, "closed": true}

func validWebLink(s string) bool {
	if s == "" {
		return true
	}
	u, e := url.Parse(s)
	return e == nil && u.Host != "" && u.User == nil && (u.Scheme == "https" || u.Scheme == "http") && len(s) <= 2000
}
func (a *App) records(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	pid := projectID(r)
	if pid < 0 || (pid > 0 && !a.projectAccess(u, pid)) {
		fail(w, 403, "This project is not assigned to you.")
		return
	}
	if r.Method == "POST" {
		a.saveRecord(w, r, u, pid)
		return
	}
	query := `SELECT r.id,COALESCE(r.project_id,0),r.author_id,u.name,r.kind,r.title,r.body,r.link,r.status,r.feedback,r.due_at,r.progress,r.score,r.criteria,COALESCE(r.assignee_id,0),COALESCE(r.parent_id,0),r.updated_at FROM workspace_records r JOIN users u ON u.id=r.author_id WHERE r.project_id=? AND r.kind!='invitation' ORDER BY r.updated_at DESC,r.id DESC`
	args := []any{pid}
	if pid == 0 {
		query = strings.Replace(query, "r.project_id=? AND r.kind!='invitation'", "r.project_id IS NULL AND (r.author_id=? OR r.status='published')", 1)
		args = []any{u.ID}
	}
	rows, e := a.db.Query(query, args...)
	if e != nil {
		fail(w, 500, "Could not load workspace.")
		return
	}
	defer rows.Close()
	out := []WorkspaceRecord{}
	for rows.Next() {
		var v WorkspaceRecord
		if rows.Scan(&v.ID, &v.ProjectID, &v.AuthorID, &v.Author, &v.Kind, &v.Title, &v.Body, &v.Link, &v.Status, &v.Feedback, &v.DueAt, &v.Progress, &v.Score, &v.Criteria, &v.AssigneeID, &v.ParentID, &v.UpdatedAt) != nil {
			fail(w, 500, "Could not read workspace.")
			return
		}
		out = append(out, v)
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read workspace.")
		return
	}
	reply(w, 200, out)
}
func (a *App) saveRecord(w http.ResponseWriter, r *http.Request, u User, pid int64) {
	var in WorkspaceRecord
	if !decode(w, r, &in) {
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	in.Link = strings.TrimSpace(in.Link)
	in.Feedback = strings.TrimSpace(in.Feedback)
	if !(pid > 0 && projectKinds[in.Kind] || pid == 0 && globalKinds[in.Kind]) || len(in.Title) < 2 || len(in.Title) > 180 || len(in.Body) > 10000 || len(in.Feedback) > 5000 || len(in.Criteria) > 3000 || !(validWebLink(in.Link) || strings.HasPrefix(in.Link, "/api/files/")) || !workspaceStatuses[in.Status] || in.Progress < 0 || in.Progress > 100 || in.Score < 0 || in.Score > 100 || in.DueAt < 0 {
		fail(w, 400, "Check the title, status, link, dates and values (0–100).")
		return
	}
	if u.Role == "student" && (!studentKinds[in.Kind] || pid == 0) {
		fail(w, 403, "Only supervisors may manage this feature.")
		return
	}
	if strings.HasPrefix(in.Link, "/api/files/") {
		fid, err := strconv.ParseInt(strings.TrimPrefix(in.Link, "/api/files/"), 10, 64)
		var fileProject int64
		if err != nil || a.db.QueryRow("SELECT project_id FROM workspace_files WHERE id=?", fid).Scan(&fileProject) != nil || fileProject != pid || pid == 0 {
			fail(w, 400, "Choose a file from this project.")
			return
		}
	}
	if in.AssigneeID > 0 {
		var n int
		a.db.QueryRow("SELECT COUNT(*) FROM users WHERE id=? AND (id=(SELECT student_id FROM projects WHERE id=?) OR id IN(SELECT user_id FROM team_members WHERE project_id=?))", in.AssigneeID, pid, pid).Scan(&n)
		if n != 1 {
			fail(w, 400, "Assignee must be a project team member.")
			return
		}
	}
	if in.ParentID > 0 {
		var n int
		a.db.QueryRow("SELECT COUNT(*) FROM workspace_records WHERE id=? AND project_id=? AND kind=?", in.ParentID, pid, in.Kind).Scan(&n)
		if n != 1 {
			fail(w, 400, "Choose an earlier version from the same project and feature.")
			return
		}
	}
	if u.Role == "teacher" && (in.Kind == "evaluation" || in.Kind == "extension") {
		var evaluate, extend bool
		a.db.QueryRow("SELECT pp.can_evaluate,pp.can_extend FROM project_permissions pp JOIN projects p ON p.id=pp.project_id WHERE pp.project_id=? AND pp.teacher_id=p.teacher_id", pid).Scan(&evaluate, &extend)
		if in.Kind == "evaluation" && !evaluate || in.Kind == "extension" && !extend {
			fail(w, 403, "The project owner must assign this responsibility in Permissions first.")
			return
		}
	}
	if in.Kind == "evaluation" && strings.TrimSpace(in.Criteria) == "" {
		fail(w, 400, "Evaluation criteria are required.")
		return
	}
	if in.Kind == "meeting" && in.DueAt == 0 {
		fail(w, 400, "Choose a meeting date and time.")
		return
	}
	tx, e := a.db.Begin()
	if e != nil {
		fail(w, 500, "Could not save workspace.")
		return
	}
	defer tx.Rollback()
	var old WorkspaceRecord
	if in.ID > 0 {
		var oldPID sql.NullInt64
		e = tx.QueryRow("SELECT id,project_id,author_id,kind,title,body,link,status,feedback,due_at,progress,score,criteria,COALESCE(assignee_id,0),COALESCE(parent_id,0),updated_at FROM workspace_records WHERE id=?", in.ID).Scan(&old.ID, &oldPID, &old.AuthorID, &old.Kind, &old.Title, &old.Body, &old.Link, &old.Status, &old.Feedback, &old.DueAt, &old.Progress, &old.Score, &old.Criteria, &old.AssigneeID, &old.ParentID, &old.UpdatedAt)
		if e != nil || oldPID.Int64 != pid || old.Kind != in.Kind || pid == 0 && old.AuthorID != u.ID || u.Role == "student" && (old.AuthorID != u.ID || old.Status != "pending" && old.Status != "changes_requested") {
			fail(w, 403, "This record is not editable by you.")
			return
		}
		if old.UpdatedAt != in.UpdatedAt {
			fail(w, 409, "This record changed. Refresh before saving.")
			return
		}
		// Reviewers keep the student's original submission intact; decisions and feedback are separate fields.
		if u.Role == "teacher" && old.AuthorID != u.ID {
			in.Title = old.Title
			in.Body = old.Body
			in.Link = old.Link
			in.AssigneeID = old.AssigneeID
			in.ParentID = old.ParentID
			if in.Kind != "meeting" && in.Kind != "extension" {
				in.DueAt = old.DueAt
			}
		}
	}
	if u.Role == "student" {
		in.Status = "pending"
		in.Feedback = ""
		in.Score = 0
		in.Criteria = ""
		in.Progress = 0
	}
	if in.Kind == "meeting" && u.Role == "teacher" && in.Status == "scheduled" {
		var clashes int
		e = tx.QueryRow(`SELECT COUNT(*) FROM workspace_records m JOIN projects p ON p.id=m.project_id WHERE m.kind='meeting' AND m.status='scheduled' AND p.teacher_id=? AND m.id!=? AND ABS(m.due_at-?)<3600`, u.ID, in.ID, in.DueAt).Scan(&clashes)
		if e != nil {
			fail(w, 500, "Could not check booking.")
			return
		}
		if clashes > 0 {
			fail(w, 409, "A meeting is already booked within one hour of this time.")
			return
		}
	}
	in.UpdatedAt = time.Now().UnixMilli()
	if in.UpdatedAt <= old.UpdatedAt {
		in.UpdatedAt = old.UpdatedAt + 1
	}
	var project, assignee, parent any
	if pid > 0 {
		project = pid
	}
	if in.AssigneeID > 0 {
		assignee = in.AssigneeID
	}
	if in.ParentID > 0 {
		parent = in.ParentID
	}
	if in.ID == 0 {
		id, err := tx.InsertID(`INSERT INTO workspace_records(project_id,author_id,kind,title,body,link,status,feedback,due_at,progress,score,criteria,assignee_id,parent_id,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, project, u.ID, in.Kind, in.Title, in.Body, in.Link, in.Status, in.Feedback, in.DueAt, in.Progress, in.Score, in.Criteria, assignee, parent, in.UpdatedAt)
		e = err
		if e == nil {
			in.ID = id
		}
		in.AuthorID = u.ID
	} else {
		in.AuthorID = old.AuthorID
		_, e = tx.Exec(`UPDATE workspace_records SET title=?,body=?,link=?,status=?,feedback=?,due_at=?,progress=?,score=?,criteria=?,assignee_id=?,parent_id=?,updated_at=? WHERE id=?`, in.Title, in.Body, in.Link, in.Status, in.Feedback, in.DueAt, in.Progress, in.Score, in.Criteria, assignee, parent, in.UpdatedAt, in.ID)
	}
	in.ProjectID = pid
	snapshot, _ := json.Marshal(in)
	if e == nil {
		_, e = tx.Exec("INSERT INTO workspace_history(record_id,actor_id,snapshot,created_at) VALUES(?,?,?,?)", in.ID, u.ID, string(snapshot), time.Now().Unix())
	}

	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		fail(w, 500, "Could not save workspace record.")
		return
	}
	if pid > 0 {
		a.notifyProject(pid, u.ID, in.Kind, in.Title+" · "+strings.ReplaceAll(in.Status, "_", " "))
	}
	reply(w, 200, in)
}
func (a *App) recordHistory(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	rid := projectID(r)
	var pid sql.NullInt64
	var owner int64
	if a.db.QueryRow("SELECT project_id,author_id FROM workspace_records WHERE id=?", rid).Scan(&pid, &owner) != nil {
		fail(w, 404, "Record unavailable.")
		return
	}
	if pid.Valid && !a.projectAccess(u, pid.Int64) || !pid.Valid && owner != u.ID {
		fail(w, 403, "History is private to this workspace.")
		return
	}
	rows, e := a.db.Query("SELECT h.snapshot,u.name,h.created_at FROM workspace_history h JOIN users u ON u.id=h.actor_id WHERE h.record_id=? ORDER BY h.id DESC", rid)
	if e != nil {
		fail(w, 500, "Could not load versions.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var s, name string
		var at int64
		if rows.Scan(&s, &name, &at) != nil {
			fail(w, 500, "Could not read history.")
			return
		}
		var v WorkspaceRecord
		if json.Unmarshal([]byte(s), &v) != nil {
			fail(w, 500, "Could not read version.")
			return
		}
		out = append(out, map[string]any{"record": v, "actor": name, "created_at": at})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read history.")
		return
	}
	reply(w, 200, out)
}
func (a *App) permissions(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	pid := projectID(r)
	if !a.projectAccess(u, pid) {
		fail(w, 403, "This project is not assigned to you.")
		return
	}
	var in struct {
		CanEvaluate bool `json:"can_evaluate"`
		CanExtend   bool `json:"can_extend"`
	}
	if r.Method == "POST" {
		if !decode(w, r, &in) {
			return
		}
		var owner int64
		a.db.QueryRow("SELECT student_id FROM projects WHERE id=?", pid).Scan(&owner)
		if u.ID != owner || u.Role != "student" {
			fail(w, 403, "Only the project owner can assign reviewer responsibilities.")
			return
		}
		_, e = a.db.Exec("INSERT INTO project_permissions(project_id,teacher_id,can_evaluate,can_extend) VALUES(?,(SELECT teacher_id FROM projects WHERE id=?),?,?) ON CONFLICT(project_id) DO UPDATE SET teacher_id=excluded.teacher_id,can_evaluate=excluded.can_evaluate,can_extend=excluded.can_extend", pid, pid, in.CanEvaluate, in.CanExtend)
	} else {
		e = a.db.QueryRow("SELECT pp.can_evaluate,pp.can_extend FROM project_permissions pp JOIN projects p ON p.id=pp.project_id WHERE pp.project_id=? AND pp.teacher_id=p.teacher_id", pid).Scan(&in.CanEvaluate, &in.CanExtend)
		if errors.Is(e, sql.ErrNoRows) {
			e = nil
		}
	}
	if e != nil {
		fail(w, 500, "Could not save or load permissions.")
		return
	}
	reply(w, 200, in)
}
func (a *App) notifyProject(pid, actor int64, kind, message string) {
	column := "notify_submissions"
	if kind == "meeting" || kind == "note" {
		column = "notify_meetings"
	}
	if kind == "request" || kind == "extension" || kind == "blocker" {
		column = "notify_requests"
	}
	_, e := a.db.Exec(`INSERT INTO workspace_notifications(user_id,project_id,kind,message,created_at) SELECT u.id,?,?,?,? FROM users u LEFT JOIN researcher_profiles rp ON rp.user_id=u.id WHERE u.id!=? AND (u.id=(SELECT student_id FROM projects WHERE id=?) OR u.id=(SELECT teacher_id FROM projects WHERE id=?) OR u.id IN(SELECT user_id FROM team_members WHERE project_id=?)) AND COALESCE(rp.`+column+`,1)=1`, pid, kind, message, time.Now().Unix(), actor, pid, pid, pid)
	if e != nil {
		return
	} // Saved work is never rolled back due to an optional notification.
}
func (a *App) notifications(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if r.Method == "POST" {
		_, e = a.db.Exec("UPDATE workspace_notifications SET read_at=? WHERE user_id=? AND read_at=0", time.Now().Unix(), u.ID)
		if e != nil {
			fail(w, 500, "Could not mark notifications read.")
			return
		}
		reply(w, 200, map[string]bool{"ok": true})
		return
	}
	rows, e := a.db.Query(`SELECT n.id,n.project_id,n.kind,n.message,n.read_at,n.created_at FROM workspace_notifications n JOIN projects p ON p.id=n.project_id WHERE n.user_id=? AND (p.student_id=? OR p.teacher_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=p.id AND user_id=?)) ORDER BY n.id DESC LIMIT 100`, u.ID, u.ID, u.ID, u.ID)
	if e != nil {
		fail(w, 500, "Could not load notifications.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, pid, read, at int64
		var kind, msg string
		if rows.Scan(&id, &pid, &kind, &msg, &read, &at) != nil {
			fail(w, 500, "Could not read notifications.")
			return
		}
		out = append(out, map[string]any{"id": id, "project_id": pid, "kind": kind, "message": msg, "read_at": read, "created_at": at})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read notifications.")
		return
	}
	reply(w, 200, out)
}
func (a *App) collaborationProjects(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	filter := "(p.student_id=? OR EXISTS(SELECT 1 FROM team_members WHERE project_id=p.id AND user_id=?))"
	args := []any{u.ID, u.ID}
	if u.Role == "teacher" {
		filter = "p.teacher_id=? AND p.status!='draft'"
		args = []any{u.ID}
	}
	rows, e := a.db.Query(`SELECT p.id,p.title,p.status,p.student_id,u.name,COALESCE(p.teacher_id,0) FROM projects p JOIN users u ON u.id=p.student_id WHERE `+filter+` ORDER BY p.updated_at DESC`, args...)
	if e != nil {
		fail(w, 500, "Could not load project list.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, owner, teacher int64
		var title, status, name string
		if rows.Scan(&id, &title, &status, &owner, &name, &teacher) != nil {
			fail(w, 500, "Could not read projects.")
			return
		}
		out = append(out, map[string]any{"id": id, "title": title, "status": status, "student_id": owner, "student": name, "teacher_id": teacher})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read projects.")
		return
	}
	reply(w, 200, out)
}

func migrateProfileExtras(db *Database) error {
	if db.postgres {
		return nil
	}
	for _, name := range []string{"photo", "qualification"} {
		rows, err := db.Query("PRAGMA table_info(researcher_profiles)")
		if err != nil {
			return err
		}
		found := false
		for rows.Next() {
			var cid, nn, pk int
			var col, typ string
			var def any
			if err = rows.Scan(&cid, &col, &typ, &nn, &def, &pk); err != nil {
				break
			}
			if col == name {
				found = true
			}
		}
		rows.Close()
		if err != nil {
			return err
		}
		if !found {
			if _, err = db.Exec("ALTER TABLE researcher_profiles ADD COLUMN " + name + " TEXT NOT NULL DEFAULT ''"); err != nil {
				return err
			}
		}
	}
	return nil
}
