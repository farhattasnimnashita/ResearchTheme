package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestResearchEthicsSupervisorReviewFlow(t *testing.T) {
	a, links := setup(t)
	register(t, a, "ethics-student@example.com", "student")
	verifyUser(t, a, links, "ethics-student@example.com")
	studentCookie := loginUser(t, a, "ethics-student@example.com", "student")

	register(t, a, "ethics-teacher@example.com", "teacher")
	verifyUser(t, a, links, "ethics-teacher@example.com")
	teacherCookie := loginUser(t, a, "ethics-teacher@example.com", "teacher")

	projectID := makeProject(t, a, studentCookie)
	var teacherID int64
	if err := a.db.QueryRow("SELECT id FROM users WHERE email=?", "ethics-teacher@example.com").Scan(&teacherID); err != nil {
		t.Fatal(err)
	}
	if _, err := a.db.Exec("UPDATE projects SET teacher_id=? WHERE id=?", teacherID, projectID); err != nil {
		t.Fatal(err)
	}

	payload := map[string]any{
		"action": "save", "project_id": projectID,
		"ethics_required": "Yes", "committee": "University Ethics Committee",
		"application_reference": "APP-001", "review_status": "Preparing",
		"submission_date": "", "decision_date": "", "approval_number": "",
		"expiry_date": "", "participants_involved": "Yes",
		"consent_method": "Online consent", "consent_plan": "Explain the survey and request voluntary consent.",
		"committee_conditions": "", "document_url": "",
	}
	expect(t, call(t, a, http.MethodPost, "/api/research-ethics", payload, studentCookie), 200)

	request := call(t, a, http.MethodPost, "/api/research-ethics", map[string]any{
		"action": "request_review", "project_id": projectID,
	}, studentCookie)
	expect(t, request, 200)

	teacherGet := call(t, a, http.MethodGet, "/api/research-ethics", nil, teacherCookie)
	expect(t, teacherGet, 200)
	if !strings.Contains(teacherGet.Body.String(), `"supervisor_review_status":"Pending"`) {
		t.Fatalf("teacher should see the pending request, got %s", teacherGet.Body.String())
	}

	teacherReview := call(t, a, http.MethodPost, "/api/research-ethics", map[string]any{
		"action": "review", "project_id": projectID,
		"supervisor_review_status": "Changes requested",
		"supervisor_comment":       "Clarify how consent will be recorded.",
	}, teacherCookie)
	expect(t, teacherReview, 200)

	studentGet := call(t, a, http.MethodGet, "/api/research-ethics", nil, studentCookie)
	expect(t, studentGet, 200)
	if !strings.Contains(studentGet.Body.String(), `"supervisor_review_status":"Changes requested"`) || !strings.Contains(studentGet.Body.String(), "Clarify how consent will be recorded.") {
		t.Fatalf("student should see teacher feedback, got %s", studentGet.Body.String())
	}

	// A teacher response confirms supervisor feedback only; it does not update the official committee status.
	if strings.Contains(teacherReview.Body.String(), `"review_status":"Approved"`) {
		t.Fatal("supervisor feedback must not become official committee approval")
	}
}

func TestResearchEthicsRequiresLogin(t *testing.T) {
	a, _ := setup(t)
	w := call(t, a, http.MethodGet, "/api/research-ethics", nil, nil)
	expect(t, w, 401)
	w = call(t, a, http.MethodPost, "/api/research-ethics", map[string]any{"project_id": 1}, nil)
	if w.Code != 401 && w.Code != 403 {
		t.Fatalf("unauthenticated POST should not be accepted, got %d: %s", w.Code, w.Body.String())
	}
}
