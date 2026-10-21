
package main

import (
	"encoding/json"
	"testing"
)

func TestTeacherDirectoryAuthentication(t *testing.T) {
	a, _ := setup(t)

	expect(
		t,
		call(t, a, "GET", "/api/teacher-directory", nil, nil),
		401,
	)
}

func TestTeacherDirectoryReadAndAccountMapping(t *testing.T) {
	a, links := setup(t)

	// A student session is needed to test the authenticated directory endpoint.
	register(t, a, "directory@example.invalid", "student")
	verifyUser(t, a, links, "directory@example.invalid")
	cookie := loginUser(t, a, "directory@example.invalid", "student")

	// Create a registered, verified application teacher account.
	register(t, a, "teacher@example.invalid", "teacher")
	verifyUser(t, a, links, "teacher@example.invalid")

	_, err := a.db.Exec(`
		CREATE TABLE teacher_directory (
			id TEXT,
			name TEXT,
			designation TEXT,
			institution TEXT,
			department TEXT,
			phone TEXT,
			email TEXT,
			website_url TEXT,
			photo_url TEXT,
			research_interests TEXT,
			expertise TEXT,
			accepting_students BOOLEAN,
			source TEXT,
			verified BOOLEAN
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = a.db.Exec(`
		INSERT INTO teacher_directory VALUES (
			'demo-1',
			'Demo Teacher',
			'Professor',
			'Demo University',
			'CSE',
			'+8801700000000',
			'teacher@example.invalid',
			'https://demo.example.invalid',
			'https://demo.example.invalid/photo.jpg',
			'["AI","Machine Learning"]',
			'["Python","NLP"]',
			TRUE,
			'test',
			TRUE
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	// The production directory endpoint intentionally requires PostgreSQL.
	expect(
		t,
		call(t, a, "GET", "/api/teacher-directory", nil, cookie),
		503,
	)

	// Read the SQLite fixture directly to test the shared directory reader.
	rows, err := a.db.DB.Query(`SELECT * FROM teacher_directory`)
	if err != nil {
		t.Fatal(err)
	}

	entries, err := readTeacherDirectory(rows)
	rows.Close()

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected one teacher, got %d", len(entries))
	}

	teacher := entries[0]

	if teacher.ID != "demo-1" ||
		teacher.Name != "Demo Teacher" ||
		teacher.Designation != "Professor" ||
		teacher.Institution != "Demo University" ||
		teacher.Department != "CSE" {
		t.Fatalf("unexpected teacher details: %+v", teacher)
	}

	if teacher.Phone == nil || *teacher.Phone != "+8801700000000" {
		t.Fatalf("unexpected phone: %+v", teacher.Phone)
	}

	if teacher.Email != "teacher@example.invalid" {
		t.Fatalf("unexpected email: %s", teacher.Email)
	}

	if teacher.WebsiteURL == nil ||
		*teacher.WebsiteURL != "https://demo.example.invalid" {
		t.Fatalf("unexpected website: %+v", teacher.WebsiteURL)
	}

	if teacher.PhotoURL == nil ||
		*teacher.PhotoURL != "https://demo.example.invalid/photo.jpg" {
		t.Fatalf("unexpected photo URL: %+v", teacher.PhotoURL)
	}

	if len(teacher.ResearchInterests) != 2 ||
		teacher.ResearchInterests[0] != "AI" ||
		teacher.ResearchInterests[1] != "Machine Learning" {
		t.Fatalf("unexpected research interests: %+v", teacher.ResearchInterests)
	}

	if len(teacher.Expertise) != 2 ||
		teacher.Expertise[0] != "Python" ||
		teacher.Expertise[1] != "NLP" {
		t.Fatalf("unexpected expertise: %+v", teacher.Expertise)
	}

	if teacher.AcceptingStudents == nil || !*teacher.AcceptingStudents {
		t.Fatalf("expected accepting_students to be true")
	}

	if !teacher.Verified || teacher.Source != "test" {
		t.Fatalf("unexpected source/verification: %+v", teacher)
	}

	// The directory email must match a verified app teacher account.
	accountRows, err := a.db.DB.Query(`
		SELECT id, email
		FROM users
		WHERE role = 'teacher' AND verified = 1
	`)
	if err != nil {
		t.Fatal(err)
	}
if err := attachRegisteredTeacherAccounts(entries, accountRows); err != nil {
    t.Fatal(err)
}

// Read the updated entry after account mapping.
teacher = entries[0]

if teacher.AccountID == nil {
    t.Fatal("expected the registered teacher account to be linked")
}

	var expectedID int64
	if err := a.db.QueryRow(
		"SELECT id FROM users WHERE email = ?",
		"teacher@example.invalid",
	).Scan(&expectedID); err != nil {
		t.Fatal(err)
	}

	if *teacher.AccountID != expectedID {
		t.Fatalf(
			"account ID mismatch: got %d, want %d",
			*teacher.AccountID,
			expectedID,
		)
	}

	data, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}

	if !json.Valid(data) {
		t.Fatal("teacher directory returned invalid JSON")
	}
}
