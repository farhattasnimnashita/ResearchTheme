package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPasswordRecoveryLifecycle(t *testing.T) {
	a, links := setup(t)
	register(t, a, "recover@example.com", "student")
	verifyUser(t, a, links, "recover@example.com")
	session := loginUser(t, a, "recover@example.com", "student")
	delivered := make(chan string, 4)
	a.send = func(to, link string) error { delivered <- link; return nil }
	w := call(t, a, "POST", "/api/forgot-password", map[string]string{"email": "recover@example.com"}, nil)
	expect(t, w, 200)
	unknown := call(t, a, "POST", "/api/forgot-password", map[string]string{"email": "unknown@example.com"}, nil)
	expect(t, unknown, 200)
	if w.Body.String() != unknown.Body.String() {
		t.Fatal("recovery reveals account existence")
	}
	var link string
	select {
	case link = <-delivered:
	case <-time.After(3 * time.Second):
		t.Fatal("email not queued")
	}
	parsed, _ := url.Parse(link)
	secret := parsed.Query().Get("token")
	if !strings.Contains(link, "/reset-password?") || len(secret) != 64 {
		t.Fatal("invalid recovery link")
	}
	var stored string
	if a.db.QueryRow("SELECT hash FROM password_resets").Scan(&stored) != nil || stored == secret || stored != digest(secret) {
		t.Fatal("token not hashed at rest")
	}
	expect(t, call(t, a, "POST", "/api/forgot-password", map[string]string{"email": "recover@example.com"}, nil), 200)
	var count int
	a.db.QueryRow("SELECT COUNT(*) FROM password_resets").Scan(&count)
	if count != 1 {
		t.Fatal("per-address cooldown failed")
	}
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"token": secret, "password": "short"}, nil), 400)
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"token": secret, "password": "NewPassword123!"}, nil), 200)
	expect(t, call(t, a, "GET", "/api/me", nil, session), 401)
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"token": secret, "password": "AnotherPass123"}, nil), 400)
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: "recover@example.com", Role: "student", Password: "StrongPass123"}, nil), 401)
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: "recover@example.com", Role: "student", Password: "NewPassword123!"}, nil), 200)
	var uid int64
	a.db.QueryRow("SELECT id FROM users WHERE email='recover@example.com'").Scan(&uid)
	a.db.Exec("INSERT INTO password_resets(hash,user_id,expires) VALUES(?,?,?)", digest(strings.Repeat("a", 64)), uid, time.Now().Unix()-1)
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"token": strings.Repeat("a", 64), "password": "NewPassword456"}, nil), 400)
	expect(t, call(t, a, "GET", "/forgot-password", nil, nil), 200)
	expect(t, call(t, a, "GET", "/reset-password", nil, nil), 200)
}
func workspaceFixture(t *testing.T) (*App, *http.Cookie, *http.Cookie, *http.Cookie, int64) {
	t.Helper()
	a, links := setup(t)
	for _, v := range []struct{ email, role string }{{"owner@example.com", "student"}, {"supervisor@example.com", "teacher"}, {"outsider@example.com", "teacher"}} {
		register(t, a, v.email, v.role)
		verifyUser(t, a, links, v.email)
	}
	owner := loginUser(t, a, "owner@example.com", "student")
	teacher := loginUser(t, a, "supervisor@example.com", "teacher")
	outsider := loginUser(t, a, "outsider@example.com", "teacher")
	id := makeProject(t, a, owner)
	var tid int64
	a.db.QueryRow("SELECT id FROM users WHERE email='supervisor@example.com'").Scan(&tid)
	expect(t, call(t, a, "POST", "/api/projects/"+itoa(id)+"/submit", map[string]any{"teacher_id": tid, "abstract": "A substantial research idea about water quality in campus."}, owner), 200)
	return a, owner, teacher, outsider, id
}
func itoa(id int64) string { return strconv.FormatInt(id, 10) }
func saveRecordTest(t *testing.T, a *App, c *http.Cookie, pid int64, v WorkspaceRecord) WorkspaceRecord {
	t.Helper()
	w := call(t, a, "POST", "/api/workspace/"+itoa(pid)+"/records", v, c)
	expect(t, w, 200)
	var out WorkspaceRecord
	if json.Unmarshal(w.Body.Bytes(), &out) != nil {
		t.Fatal("invalid record")
	}
	return out
}
func TestWorkspaceAuthorizationAndVersionReviews(t *testing.T) {
	a, owner, teacher, outsider, id := workspaceFixture(t)
	path := "/api/workspace/" + itoa(id)
	expect(t, call(t, a, "GET", path+"/records", nil, nil), 401)
	expect(t, call(t, a, "GET", path+"/records", nil, outsider), 403)
	expect(t, call(t, a, "POST", path+"/records", WorkspaceRecord{Kind: "milestone", Title: "First stage", Status: "approved"}, owner), 403)
	submission := saveRecordTest(t, a, owner, id, WorkspaceRecord{Kind: "submission", Title: "First report", Body: "Original student report", Status: "approved", Score: 100, Feedback: "forged"})
	if submission.Status != "pending" || submission.Score != 0 || submission.Feedback != "" {
		t.Fatal("student self-approved work")
	}
	expect(t, call(t, a, "POST", path+"/records", submission, outsider), 403)
	submission.Status = "changes_requested"
	submission.Feedback = "Explain the experiment controls"
	submission.Body = "malicious reviewer rewrite"
	reviewed := saveRecordTest(t, a, teacher, id, submission)
	if reviewed.Body != "Original student report" || reviewed.Feedback == "" {
		t.Fatal("review destroyed student evidence")
	}
	expect(t, call(t, a, "POST", path+"/records", submission, teacher), 409)
	expect(t, call(t, a, "GET", "/api/records/"+itoa(reviewed.ID)+"/history", nil, outsider), 403)
	w := call(t, a, "GET", "/api/records/"+itoa(reviewed.ID)+"/history", nil, owner)
	expect(t, w, 200)
	var versions []any
	json.Unmarshal(w.Body.Bytes(), &versions)
	if len(versions) != 2 {
		t.Fatal("version snapshots missing")
	}
	revision := saveRecordTest(t, a, owner, id, WorkspaceRecord{Kind: "submission", Title: "Report v2", Body: "Revised experiment controls", Status: "pending", ParentID: submission.ID})
	if revision.ID == submission.ID || revision.ParentID != submission.ID {
		t.Fatal("submission version replaced original")
	}
	expect(t, call(t, a, "POST", path+"/records", WorkspaceRecord{Kind: "note", Title: "Unsafe link", Link: "javascript:alert(1)", Status: "active"}, teacher), 400)
	expect(t, call(t, a, "POST", path+"/records", WorkspaceRecord{Kind: "note", Title: "Wrong parent", ParentID: submission.ID, Status: "active"}, teacher), 400)
	expect(t, call(t, a, "POST", "/api/workspace/0/records", WorkspaceRecord{Kind: "opportunity", Title: "Student advert", Status: "published"}, owner), 403)
	advert := saveRecordTest(t, a, teacher, 0, WorkspaceRecord{Kind: "opportunity", Title: "Paid research assistant", Body: "Paid: 5000 BDT per month. Apply by email.", Status: "active"})
	w = call(t, a, "GET", "/api/workspace/0/records", nil, owner)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), advert.Title) {
		t.Fatal("draft advertisement leaked")
	}
	advert.Status = "published"
	saveRecordTest(t, a, teacher, 0, advert)
	w = call(t, a, "GET", "/api/workspace/0/records", nil, owner)
	if !strings.Contains(w.Body.String(), advert.Title) {
		t.Fatal("published opportunity missing")
	}
	expect(t, call(t, a, "POST", "/api/workspace/0/records", advert, outsider), 403)
}
func TestReviewerResponsibilitiesAndMeetingConflicts(t *testing.T) {
	a, owner, teacher, outsider, id := workspaceFixture(t)
	path := "/api/workspace/" + itoa(id)
	evaluation := WorkspaceRecord{Kind: "evaluation", Title: "Method evaluation", Status: "completed", Criteria: "Clarity 30; method 40; validation 30", Score: 78}
	expect(t, call(t, a, "POST", path+"/records", evaluation, teacher), 403)
	expect(t, call(t, a, "POST", path+"/permissions", map[string]bool{"can_evaluate": true, "can_extend": true}, teacher), 403)
	expect(t, call(t, a, "POST", path+"/permissions", map[string]bool{"can_evaluate": true, "can_extend": true}, owner), 200)
	saveRecordTest(t, a, teacher, id, evaluation)
	saveRecordTest(t, a, teacher, id, WorkspaceRecord{Kind: "extension", Title: "Deadline extension", DueAt: time.Now().Add(24 * time.Hour).Unix(), Status: "approved", Feedback: "Two extra days allowed"})
	due := time.Now().Add(48 * time.Hour).Unix()
	saveRecordTest(t, a, teacher, id, WorkspaceRecord{Kind: "meeting", Title: "Weekly research discussion", DueAt: due, Status: "scheduled"})
	expect(t, call(t, a, "POST", path+"/records", WorkspaceRecord{Kind: "meeting", Title: "Conflicting meeting", DueAt: due + 1800, Status: "scheduled"}, teacher), 409)
	saveRecordTest(t, a, teacher, id, WorkspaceRecord{Kind: "meeting", Title: "Following meeting", DueAt: due + 3600, Status: "scheduled"})
	expect(t, call(t, a, "GET", path+"/permissions", nil, outsider), 403)
}
func TestNotificationsAndProfilePreferences(t *testing.T) {
	a, owner, teacher, _, id := workspaceFixture(t)
	expect(t, call(t, a, "POST", "/api/researcher-profile", ResearcherProfile{Bio: "Supervisor working on systems", Interests: "Security", Expertise: "Go", NotifyRequests: true, NotifyMeetings: false, NotifySubmissions: true}, teacher), 200)
	saveRecordTest(t, a, owner, id, WorkspaceRecord{Kind: "meeting", Title: "Meeting request", DueAt: time.Now().Add(time.Hour).Unix(), Status: "pending"})
	saveRecordTest(t, a, owner, id, WorkspaceRecord{Kind: "submission", Title: "Research report", Body: "Report body", Status: "pending"})
	w := call(t, a, "GET", "/api/notifications", nil, teacher)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), "Meeting request") || !strings.Contains(w.Body.String(), "Research report") {
		t.Fatal("notification preferences not respected")
	}
	expect(t, call(t, a, "POST", "/api/notifications", map[string]any{}, teacher), 200)
	w = call(t, a, "GET", "/api/notifications", nil, owner)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), "Research report") {
		t.Fatal("self notification generated")
	}
}
func TestOptInTeamMembershipAndProtectedFiles(t *testing.T) {
	a, owner, teacher, outsider, id := workspaceFixture(t)
	path := "/api/workspace/" + itoa(id)
	var mateID int64
	_, e := a.db.Exec("INSERT INTO users(name,email,password_hash,role,verified,created_at) VALUES('Team Mate','mate@example.com','unused','student',1,?)", time.Now().Unix())
	if e != nil {
		t.Fatal(e)
	}
        e = a.db.QueryRow(
                "SELECT id FROM users WHERE email=?",
                "mate@example.com",
        ).Scan(&mateID)
        if e != nil {
                t.Fatal("could not find mate user:", e)
        }
	raw := "mate-session"
	a.db.Exec("INSERT INTO sessions(hash,user_id,expires) VALUES(?,?,?)", digest(raw), mateID, time.Now().Add(time.Hour).Unix())
	mate := &http.Cookie{Name: "rb_session", Value: raw}
	expect(t, call(t, a, "GET", path+"/members", nil, mate), 403)
	expect(t, call(t, a, "POST", path+"/members", map[string]string{"email": "mate@example.com", "action": "invite"}, teacher), 403)
	expect(t, call(t, a, "POST", path+"/members", map[string]string{"email": "mate@example.com", "action": "invite"}, owner), 201)
	w := call(t, a, "GET", "/api/invitations", nil, mate)
	expect(t, w, 200)
	var invites []struct{ ID int64 }
	json.Unmarshal(w.Body.Bytes(), &invites)
	if len(invites) != 1 {
		t.Fatal("invitation missing")
	}
	expect(t, call(t, a, "POST", "/api/invitations", map[string]any{"id": invites[0].ID, "accept": true}, mate), 200)
	expect(t, call(t, a, "GET", path+"/members", nil, mate), 200)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	f, _ := writer.CreateFormFile("file", "report.txt")
	f.Write([]byte("Research result version one"))
	writer.Close()
	req := httptest.NewRequest("POST", path+"/files", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Origin", a.cfg.URL)
	req.Header.Set("X-Requested-With", "ResearchBridge")
	req.AddCookie(owner)
	req.RemoteAddr = "127.0.0.2:1234"
	rr := httptest.NewRecorder()
	a.routes().ServeHTTP(rr, req)
	expect(t, rr, 201)
	var upload struct {
		ID   int64
		Link string
	}
	json.Unmarshal(rr.Body.Bytes(), &upload)
	expect(t, call(t, a, "GET", upload.Link, nil, outsider), 403)
	expect(t, call(t, a, "GET", upload.Link, nil, nil), 401)
	w = call(t, a, "GET", upload.Link, nil, teacher)
	expect(t, w, 200)
	if !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") || w.Body.String() != "Research result version one" {
		t.Fatal("protected attachment failed")
	}
	saveRecordTest(t, a, mate, id, WorkspaceRecord{Kind: "submission", Title: "Team report", Link: upload.Link, Status: "pending"})
	expect(t, call(t, a, "POST", path+"/records", WorkspaceRecord{Kind: "submission", Title: "Wrong attachment", Link: "/api/files/999999", Status: "pending"}, owner), 400)
}
func TestResetEmailContent(t *testing.T) {
	subject, body := accountMailText("http://localhost:8080/reset-password?token=abc")
	if !strings.Contains(subject, "Reset") || !strings.Contains(body, "30 minutes") || strings.Contains(subject, "Verify") {
		t.Fatal("reset email uses wrong purpose")
	}
	b, e := verificationMail("sender@gmail.com", "recipient@example.com", "http://localhost:8080/reset-password?token=abc")
	if e != nil || !strings.Contains(string(b), "Subject: Reset your ResearchBridge password") {
		t.Fatal("SMTP reset message incorrect", e)
	}
}

func TestResponsibilitiesDoNotTransferToAnotherSupervisor(t *testing.T) {
	a, owner, supervisor, outsider, id := workspaceFixture(t)
	path := "/api/workspace/" + itoa(id)
	expect(t, call(t, a, "POST", path+"/permissions", map[string]bool{"can_evaluate": true, "can_extend": true}, owner), 200)
	var other int64
	a.db.QueryRow("SELECT id FROM users WHERE email='outsider@example.com'").Scan(&other)
	a.db.Exec("UPDATE projects SET teacher_id=? WHERE id=?", other, id)
	expect(t, call(t, a, "GET", path+"/records", nil, supervisor), 403)
	evaluation := WorkspaceRecord{Kind: "evaluation", Title: "New reviewer evaluation", Status: "completed", Criteria: "Reasoning and evidence", Score: 80}
	expect(t, call(t, a, "POST", path+"/records", evaluation, outsider), 403)
	expect(t, call(t, a, "POST", path+"/permissions", map[string]bool{"can_evaluate": true, "can_extend": false}, owner), 200)
	saveRecordTest(t, a, outsider, id, evaluation)
}
func TestAllResearchModulesPersistAndMigrationRepeats(t *testing.T) {
	a, owner, supervisor, _, id := workspaceFixture(t)
	expect(t, call(t, a, "POST", "/api/workspace/"+itoa(id)+"/permissions", map[string]bool{"can_evaluate": true, "can_extend": true}, owner), 200)
	for kind := range projectKinds {
		// This test exercises the full module catalogue, not the shared API rate limit.
		a.limits = nil
		record := WorkspaceRecord{Kind: kind, Title: "Saved " + kind + " record", Body: "Research notes and review details", Status: "active", Criteria: "Evidence and method", Score: 60, Progress: 25}
		if kind == "meeting" {
			record.DueAt = time.Now().Add(24 * time.Hour).Unix()
		}
		saveRecordTest(t, a, supervisor, id, record)
	}
	for kind := range globalKinds {
		a.limits = nil
		saveRecordTest(t, a, supervisor, 0, WorkspaceRecord{Kind: kind, Title: "Published " + kind, Body: "Members, projects, details and application instructions", Status: "published"})
	}
	if !a.db.postgres {
		if err := migrateWorkspace(a.db.DB); err != nil {
			t.Fatal(err)
		}
	}
	var n int
	a.db.QueryRow("SELECT COUNT(*) FROM workspace_records WHERE project_id=?", id).Scan(&n)
	if n != len(projectKinds) {
		t.Fatal("migration deleted module records")
	}
	w := call(t, a, "GET", "/api/workspace/"+itoa(id)+"/records", nil, owner)
	expect(t, w, 200)
	var records []WorkspaceRecord
	json.Unmarshal(w.Body.Bytes(), &records)
	if len(records) != len(projectKinds) {
		t.Fatal("student cannot read saved project modules")
	}
}
