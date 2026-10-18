package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Project struct {
	TeacherID int64  `json:"teacher_id"`
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Area      string `json:"area"`
	Abstract  string `json:"abstract"`
	Status    string `json:"status"`
	Feedback  string `json:"feedback"`
	Student   string `json:"student"`
	Teacher   string `json:"teacher"`
	Updated   int64  `json:"updated"`
}

func (a *App) teachers(w http.ResponseWriter, r *http.Request) {
	if _, e := a.user(r); e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	rows, e := a.db.Query("SELECT id,name FROM users WHERE role='teacher' AND verified=1 ORDER BY name")
	if e != nil {
		fail(w, 500, "Could not load supervisors.")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var name string
		if rows.Scan(&id, &name) != nil {
			fail(w, 500, "Could not read supervisors.")
			return
		}
		out = append(out, map[string]any{"id": id, "name": name})
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read supervisors.")
		return
	}
	reply(w, 200, out)
}
func (a *App) listProjects(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	filter := "p.student_id=?"
	if u.Role == "teacher" {
		filter = "p.teacher_id=? AND p.status!='draft'"
	}
	rows, e := a.db.Query(`SELECT p.id,p.title,p.area,p.abstract,p.status,p.feedback,s.name,COALESCE(t.name,'Not assigned'),COALESCE(t.id,0),p.updated_at FROM projects p JOIN users s ON p.student_id=s.id LEFT JOIN users t ON p.teacher_id=t.id WHERE `+filter+` ORDER BY p.updated_at DESC,p.id DESC`, u.ID)
	if e != nil {
		fail(w, 500, "Could not load projects.")
		return
	}
	defer rows.Close()
	out := []Project{}
	for rows.Next() {
		var p Project
		if rows.Scan(&p.ID, &p.Title, &p.Area, &p.Abstract, &p.Status, &p.Feedback, &p.Student, &p.Teacher, &p.TeacherID, &p.Updated) != nil {
			fail(w, 500, "Could not read projects.")
			return
		}
		out = append(out, p)
	}
	if rows.Err() != nil {
		fail(w, 500, "Could not read projects.")
		return
	}
	reply(w, 200, out)
}
func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Only students can create research proposals.")
		return
	}
	var p struct {
		Title     string `json:"title"`
		Area      string `json:"area"`
		Abstract  string `json:"abstract"`
		TeacherID int64  `json:"teacher_id"`
	}
	if !decode(w, r, &p) {
		return
	}
	p.Title = strings.TrimSpace(p.Title)
	p.Area = strings.TrimSpace(p.Area)
	p.Abstract = strings.TrimSpace(p.Abstract)
	if len(p.Title) < 5 || len(p.Title) > 160 || len(p.Area) < 2 || len(p.Area) > 80 || len(p.Abstract) < 20 || len(p.Abstract) > 6000 {
		fail(w, 400, "Use a title (5–160 characters), research area (2–80), and abstract (20–6000).")
		return
	}
	var teacher any
	if p.TeacherID != 0 {
		var id int64
		if a.db.QueryRow("SELECT id FROM users WHERE id=? AND role='teacher' AND verified=1", p.TeacherID).Scan(&id) != nil {
			fail(w, 400, "Select a verified teacher.")
			return
		}
		teacher = p.TeacherID
	}
	id, e := a.db.InsertID("INSERT INTO projects(student_id,teacher_id,title,area,abstract,updated_at) VALUES(?,?,?,?,?,?)", u.ID, teacher, p.Title, p.Area, p.Abstract, time.Now().Unix())
	if e != nil {
		fail(w, 500, "Could not save proposal.")
		return
	}
	reply(w, 201, map[string]any{"id": id, "message": "Draft saved. Submit it when you are ready for review."})
}
func (a *App) submitProject(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "student" {
		fail(w, 403, "Only students can submit.")
		return
	}
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		fail(w, 400, "Invalid proposal.")
		return
	}
	var p struct {
		TeacherID int64  `json:"teacher_id"`
		Abstract  string `json:"abstract"`
	}
	if !decode(w, r, &p) {
		return
	}
	p.Abstract = strings.TrimSpace(p.Abstract)
	if p.TeacherID <= 0 || len(p.Abstract) < 20 || len(p.Abstract) > 6000 {
		fail(w, 400, "Choose a supervisor and enter an abstract (20–6000 characters).")
		return
	}
	result, e := a.db.Exec(`UPDATE projects SET teacher_id=?,abstract=?,status='submitted',updated_at=? WHERE id=? AND student_id=? AND status IN ('draft','changes_requested') AND EXISTS(SELECT 1 FROM users WHERE id=? AND role='teacher' AND verified=1)`, p.TeacherID, p.Abstract, time.Now().Unix(), id, u.ID, p.TeacherID)
	if e != nil {
		fail(w, 500, "Could not submit proposal.")
		return
	}
	n, e := result.RowsAffected()
	if e != nil || n == 0 {
		fail(w, 409, "Proposal unavailable, supervisor invalid, or proposal is no longer editable.")
		return
	}
	a.notifyProject(id, u.ID, "request", "New proposal ready for supervision: review requested")
	reply(w, 200, map[string]string{"message": "Proposal submitted to your supervisor."})
}
func (a *App) reviewProject(w http.ResponseWriter, r *http.Request) {
	u, e := a.user(r)
	if e != nil {
		fail(w, 401, "Please log in.")
		return
	}
	if u.Role != "teacher" {
		fail(w, 403, "Only teachers can review.")
		return
	}
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if e != nil {
		fail(w, 400, "Invalid proposal.")
		return
	}
	var p struct {
		Status   string `json:"status"`
		Feedback string `json:"feedback"`
	}
	if !decode(w, r, &p) {
		return
	}
	p.Feedback = strings.TrimSpace(p.Feedback)
	if (p.Status != "approved" && p.Status != "changes_requested" && p.Status != "rejected") || len(p.Feedback) < 5 || len(p.Feedback) > 3000 {
		fail(w, 400, "Choose a decision and add feedback (5–3000 characters).")
		return
	}
	result, e := a.db.Exec("UPDATE projects SET status=?,feedback=?,updated_at=? WHERE id=? AND teacher_id=? AND status='submitted'", p.Status, p.Feedback, time.Now().Unix(), id, u.ID)
	if e != nil {
		fail(w, 500, "Could not save review.")
		return
	}
	n, e := result.RowsAffected()
	if e != nil || n == 0 {
		fail(w, 409, "Proposal unavailable or already reviewed.")
		return
	}
	a.notifyProject(id, u.ID, "submission", "Proposal review: "+p.Status)
	reply(w, 200, map[string]string{"message": "Review saved and shared with the student."})
}
