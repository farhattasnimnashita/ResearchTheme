package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestStudentAndTeacherSignupProfiles(t *testing.T) {
	a, links := setup(t)
	student := credentials{Name: "Student One", Email: "profile-student@example.com", Password: "StrongPass123", Role: "student", Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}
	expect(t, call(t, a, "POST", "/api/register", student, nil), 201)
	verifyUser(t, a, links, student.Email)
	cookie := loginUser(t, a, student.Email, student.Role)
	w := call(t, a, "GET", "/api/me", nil, cookie)
	expect(t, w, 200)
	var profile User
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.Institution != "SUST" || profile.Department != "CSE" || profile.AcademicYear != 2 || profile.Semester != 4 {
		t.Fatal("student details were not saved", profile)
	}
	teacher := credentials{Name: "Teacher One", Email: "profile-teacher@example.com", Password: "StrongPass123", Role: "teacher", Institution: "University of Dhaka", Department: "should not be stored", AcademicYear: 6, Semester: 12}
	expect(t, call(t, a, "POST", "/api/register", teacher, nil), 201)
	verifyUser(t, a, links, teacher.Email)
	cookie = loginUser(t, a, teacher.Email, teacher.Role)
	w = call(t, a, "GET", "/api/me", nil, cookie)
	expect(t, w, 200)
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil {
		t.Fatal(err)
	}
	if profile.Institution != "University of Dhaka" || profile.Department != "" || profile.AcademicYear != 0 || profile.Semester != 0 {
		t.Fatal("teacher saved student-only details", profile)
	}
}
func TestSignupRequiresRoleSpecificDetails(t *testing.T) {
	for _, which := range []string{"institution", "department", "year", "semester"} {
		t.Run(which, func(t *testing.T) {
			a, _ := setup(t)
			c := credentials{Name: "Student One", Email: "invalid@example.com", Password: "StrongPass123", Role: "student", Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}
			switch which {
			case "institution":
				c.Institution = " "
			case "department":
				c.Department = ""
			case "year":
				c.AcademicYear = 0
			case "semester":
				c.Semester = 13
			}
			expect(t, call(t, a, "POST", "/api/register", c, nil), 400)
			var n int
			if err := a.db.QueryRow("SELECT count(*) FROM users").Scan(&n); err != nil || n != 0 {
				t.Fatal("invalid signup wrote an account", err)
			}
		})
	}
	a, _ := setup(t)
	c := credentials{Name: "Teacher One", Email: "invalid-teacher@example.com", Password: "StrongPass123", Role: "teacher"}
	expect(t, call(t, a, "POST", "/api/register", c, nil), 400)
}
func TestLoginSuggestsSignupOnlyForMissingAccount(t *testing.T) {
	a, links := setup(t)
	w := call(t, a, "POST", "/api/login", credentials{Email: "missing@example.com", Password: "StrongPass123", Role: "student"}, nil)
	expect(t, w, 404)
	if !strings.Contains(w.Body.String(), `"suggest_signup":true`) || !strings.Contains(w.Body.String(), "account_not_found") {
		t.Fatal("missing signup guidance")
	}
	register(t, a, "existing@example.com", "student")
	verifyUser(t, a, links, "existing@example.com")
	for _, test := range []struct {
		role, password, code string
		status               int
	}{{"student", "wrong", "invalid_password", 401}, {"teacher", "StrongPass123", "role_mismatch", 403}} {
		w = call(t, a, "POST", "/api/login", credentials{Email: "existing@example.com", Password: test.password, Role: test.role}, nil)
		expect(t, w, test.status)
		if !strings.Contains(w.Body.String(), test.code) || strings.Contains(w.Body.String(), "suggest_signup") {
			t.Fatal("existing account incorrectly suggested signup", w.Body.String())
		}
	}
}
func TestProfileMigrationPreservesExistingData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE users(id INTEGER PRIMARY KEY,name TEXT NOT NULL,email TEXT NOT NULL UNIQUE,password_hash TEXT NOT NULL,role TEXT NOT NULL,verified INTEGER NOT NULL DEFAULT 0,created_at INTEGER NOT NULL);
 INSERT INTO users(id,name,email,password_hash,role,verified,created_at) VALUES(1,'Existing Student','old@example.com','existing_hash','student',1,1);`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	db, err = openSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO sessions(hash,user_id,expires) VALUES('saved-session',1,9999999999); INSERT INTO projects(student_id,title,area,abstract,updated_at) VALUES(1,'Existing proposal','NLP','An existing abstract preserved across migration.',1);`)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	db, err = openSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var name, hash, institution, department string
	var year, semester int
	var verified bool
	err = db.QueryRow("SELECT name,password_hash,verified,institution,department,academic_year,semester FROM users WHERE id=1").Scan(&name, &hash, &verified, &institution, &department, &year, &semester)
	if err != nil || name != "Existing Student" || hash != "existing_hash" || !verified || institution != "" || department != "" || year != 0 || semester != 0 {
		t.Fatal("old account altered or profile migration failed", err)
	}
	for _, table := range []string{"sessions", "projects"} {
		var count int
		if err = db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			t.Fatal("old data lost", table, err)
		}
	}
	if err = migrateUserProfile(db); err != nil {
		t.Fatal("migration was not idempotent", err)
	}
}
func TestSMTPResendMessageDoesNotMentionConsole(t *testing.T) {
	a, _ := setup(t)
	a.cfg.EmailMode = "smtp"
	w := call(t, a, http.MethodPost, "/api/resend", map[string]string{"email": "unknown@example.com"}, nil)
	expect(t, w, 200)
	if strings.Contains(w.Body.String(), "console") || strings.Contains(w.Body.String(), "terminal") {
		t.Fatal("SMTP still displayed console instructions")
	}
}
