package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSQLBinding(t *testing.T) {
	q := `SELECT '?', "?", 'it''s ?', ? FROM users WHERE id=?`
	want := `SELECT '?', "?", 'it''s ?', $1 FROM users WHERE id=$2`
	if got := bindSQL(q, true); got != want {
		t.Fatalf("binding: %s", got)
	}
	if bindSQL(q, false) != q {
		t.Fatal("SQLite SQL changed")
	}
}
func TestInvalidPostgresDoesNotFallBack(t *testing.T) {
	db, err := openConfiguredDB(Config{DatabaseURL: "invalid-database-url", DBPath: "must-not-create.db"})
	if err == nil {
		db.Close()
		t.Fatal("invalid PostgreSQL URL accepted")
	}
}
func TestDirectorySearchAndProfileFlags(t *testing.T) {
	a, links := setup(t)
	register(t, a, "searcher@example.com", "student")
	verifyUser(t, a, links, "searcher@example.com")
	register(t, a, "result@example.com", "student")
	verifyUser(t, a, links, "result@example.com")
	c := loginUser(t, a, "searcher@example.com", "student")
	if _, err := a.db.Exec("UPDATE users SET name=? WHERE email=?", "Student 100%", "result@example.com"); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"Student", "100%25"} {
		w := call(t, a, "GET", "/api/students?q="+query, nil, c)
		expect(t, w, 200)
		var students []DirectoryStudent
		if err := json.Unmarshal(w.Body.Bytes(), &students); err != nil || len(students) != 1 {
			t.Fatalf("search failed: %s", w.Body.String())
		}
	}
	w := call(t, a, "POST", "/api/researcher-profile", ResearcherProfile{Bio: "Updated", NotifyRequests: true, NotifySubmissions: false, NotifyMeetings: true}, c)
	expect(t, w, 200)
	w = call(t, a, "GET", "/api/researcher-profile", nil, c)
	expect(t, w, 200)
	var profile ResearcherProfile
	if err := json.Unmarshal(w.Body.Bytes(), &profile); err != nil || !profile.NotifyRequests || profile.NotifySubmissions || !profile.NotifyMeetings {
		t.Fatalf("flags failed: %s", w.Body.String())
	}
}

func TestPostgresMigrationPreservesData(t *testing.T) {
	connection := os.Getenv("POSTGRES_TEST_URL")
	if connection == "" {
		t.Skip("POSTGRES_TEST_URL not set")
	}
	schema := fmt.Sprintf("rb_test_%d", time.Now().UnixNano())
	db, err := openPostgresSchema(connection, schema)
	if err != nil {
		t.Fatal(err)
	}
	id, err := db.InsertID("INSERT INTO users(name,email,password_hash,role,created_at) VALUES(?,?,?,?,?)", "Keep Student", "keep@example.com", "hash", "student", int64(1))
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	db.Close()
	db, err = openPostgresSchema(connection, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
	var name string
	if err = db.QueryRow("SELECT name FROM users WHERE id=?", id).Scan(&name); err != nil || name != "Keep Student" {
		t.Fatalf("restart lost data: %v", err)
	}
	var actualSchema string
	if err = db.QueryRow("SELECT current_schema()").Scan(&actualSchema); err != nil || actualSchema != schema {
		t.Fatal("wrong database schema", err)
	}
}
