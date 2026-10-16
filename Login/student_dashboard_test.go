package main

import (
	"strings"
	"testing"
)

func TestResearchThemeDashboardAccess(t *testing.T) {
	a, links := setup(t)
	for _, path := range []string{"/dashboard/student", "/workspace/student"} {
		w := call(t, a, "GET", path, nil, nil)
		expect(t, w, 303)
		if w.Header().Get("Location") != "/login/student" {
			t.Fatal("student login required")
		}
	}
	register(t, a, "student@example.com", "student")
	verifyUser(t, a, links, "student@example.com")
	student := loginUser(t, a, "student@example.com", "student")
	w := call(t, a, "GET", "/dashboard/student", nil, student)
	expect(t, w, 200)
	if !strings.Contains(w.Body.String(), "/assets/researchtheme-dashboard.js") || strings.Contains(w.Body.String(), "<script>") {
		t.Fatal("must serve imported dashboard with CSP-compatible external script")
	}
	expect(t, call(t, a, "GET", "/workspace/student", nil, student), 200)
	register(t, a, "teacher@example.com", "teacher")
	verifyUser(t, a, links, "teacher@example.com")
	teacher := loginUser(t, a, "teacher@example.com", "teacher")
	for _, path := range []string{"/dashboard/student", "/workspace/student"} {
		w = call(t, a, "GET", path, nil, teacher)
		expect(t, w, 303)
		if w.Header().Get("Location") != "/dashboard/teacher" {
			t.Fatal("teacher must return to own dashboard")
		}
	}
	w = call(t, a, "GET", "/dashboard/teacher", nil, teacher)
	expect(t, w, 200)
	if !strings.Contains(w.Body.String(), "/assets/teacher-dashboard.js") {
		t.Fatal("teacher dashboard must use the sky-blue workspace")
	}
	expect(t, call(t, a, "GET", "/assets/student-dashboard.html", nil, nil), 404)
}
