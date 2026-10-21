package main

import (
	"encoding/json"
	"testing"
)

func TestTeacherDirectoryAuthentication(t *testing.T) {
	a, _ := setup(t)
	expect(t, call(t, a, "GET", "/api/teacher-directory", nil, nil), 401)
}

func TestTeacherDirectoryRead(t *testing.T) {
	a, links := setup(t)
	register(t, a, "directory@example.invalid", "student")
	verifyUser(t, a, links, "directory@example.invalid")
	cookie := loginUser(t, a, "directory@example.invalid", "student")
	// SQLite is a fixture for the SQL reader; production requires PostgreSQL.
	_, err := a.db.Exec(`CREATE TABLE teacher_directory (id TEXT, name TEXT, designation TEXT, institution TEXT, department TEXT, phone TEXT, email TEXT, website_url TEXT, photo_url TEXT, research_interests TEXT, expertise TEXT, accepting_students BOOLEAN, source TEXT, verified BOOLEAN)`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.db.Exec(`INSERT INTO teacher_directory VALUES ('demo-1','Demo Teacher','Professor','Demo University','CSE',NULL,'teacher@example.invalid',NULL,NULL,'["AI"]','[]',NULL,'test',FALSE)`)
	if err != nil {
		t.Fatal(err)
	}
	if !a.db.postgres {
		expect(t, call(t, a, "GET", "/api/teacher-directory", nil, cookie), 503)
	}
	rows, err := a.db.DB.Query(`SELECT * FROM teacher_directory`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	entries, err := readTeacherDirectory(rows)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ID != "demo-1" || entries[0].ResearchInterests[0] != "AI" || entries[0].Phone != nil || entries[0].AcceptingStudents != nil {
		t.Fatalf("unexpected records: %+v", entries)
	}
	b, err := json.Marshal(entries)
	if err != nil || !json.Valid(b) {
		t.Fatal("invalid JSON")
	}
}
