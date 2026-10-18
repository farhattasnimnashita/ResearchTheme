package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func sbSetup(t *testing.T, handler http.HandlerFunc) *App {
	t.Helper()
	a, _ := setup(t)
	a.cfg.EmailMode = "supabase"
	a.cfg.SupabaseKey = "test-public-key"
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	a.cfg.SupabaseURL = server.URL
	if err := initSupabaseTables(a.db); err != nil {
		t.Fatal(err)
	}
	return a
}
func TestSupabaseLoginLinksExistingProfile(t *testing.T) {
	a := sbSetup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/token" || r.Header.Get("apikey") != "test-public-key" {
			t.Error("unexpected auth request")
		}
		reply(w, 200, sbResponse{AccessToken: "not-exposed", ExpiresIn: 3600, User: sbUser{ID: "auth-id-1", Email: "existing@example.com", Confirmed: "2026-01-01", Metadata: credentials{Role: "teacher", Name: "Overwrite attempt"}}})
	})
	id, err := a.db.InsertID("INSERT INTO users(name,email,password_hash,role,verified,created_at) VALUES(?,?,?,?,?,?)", "Original", "existing@example.com", "oldhash", "student", 1, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	a.db.Exec("INSERT INTO sessions(hash,user_id,expires) VALUES(?,?,?)", digest("legacy"), id, time.Now().Add(time.Hour).Unix())
	expect(t, call(t, a, "GET", "/api/me", nil, &http.Cookie{Name: "rb_session", Value: "legacy"}), 401)
	res := call(t, a, "POST", "/api/login", credentials{Email: "existing@example.com", Password: "StrongPass123", Role: "student"}, nil)
	expect(t, res, 200)
	cookies := res.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly {
		t.Fatal("missing safe app cookie")
	}
	me := call(t, a, "GET", "/api/me", nil, cookies[0])
	expect(t, me, 200)
	var user User
	json.Unmarshal(me.Body.Bytes(), &user)
	if user.ID != id || user.Name != "Original" || user.Role != "student" {
		t.Fatal("profile overwritten")
	}
	var hash string
	a.db.QueryRow("SELECT password_hash FROM users WHERE id=?", id).Scan(&hash)
	if hash != "!supabase-auth" {
		t.Fatal("legacy password still enabled")
	}
	var count int
	a.db.QueryRow("SELECT count(*) FROM sessions WHERE hash=?", digest("legacy")).Scan(&count)
	if count != 0 {
		t.Fatal("legacy session remains")
	}
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: "existing@example.com", Password: "StrongPass123", Role: "teacher"}, nil), 403)
}
func TestSupabaseUnconfirmedCannotLogin(t *testing.T) {
	a := sbSetup(t, func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, sbResponse{AccessToken: "x", User: sbUser{ID: "u", Email: "a@example.com"}})
	})
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: "a@example.com", Password: "StrongPass123", Role: "student"}, nil), 403)
}
func TestSupabaseResetRequiresRecoveryAndRevokesSessions(t *testing.T) {
	updated := false
	a := sbSetup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth/v1/verify" {
			var in map[string]string
			json.NewDecoder(r.Body).Decode(&in)
			if in["type"] != "recovery" {
				t.Fatal("wrong verification purpose")
			}
			if in["token"] != "12345678" {
				http.Error(w, "bad", 400)
				return
			}
			reply(w, 200, sbResponse{AccessToken: "recovery-secret", User: sbUser{ID: "reset-user", Email: "a@example.com", Confirmed: "yes"}})
			return
		}
		if r.URL.Path == "/auth/v1/user" && r.Method == "PUT" && r.Header.Get("Authorization") == "Bearer recovery-secret" {
			updated = true
			reply(w, 200, map[string]string{})
			return
		}
		t.Error("unexpected call")
		http.Error(w, "bad", 400)
	})
	id, _ := a.db.InsertID("INSERT INTO users(name,email,password_hash,role,verified,created_at) VALUES(?,?,?,?,?,?)", "Test", "a@example.com", "!supabase-auth", "student", 1, time.Now().Unix())
	a.db.Exec("INSERT INTO supabase_identities(auth_id,user_id) VALUES(?,?)", "reset-user", id)
	a.db.Exec("INSERT INTO sessions(hash,user_id,expires) VALUES(?,?,?)", "session", id, time.Now().Add(time.Hour).Unix())
	a.db.Exec("INSERT INTO supabase_sessions(hash) VALUES(?)", "session")
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"email": "a@example.com", "token": "00000000", "password": "NewStrong123"}, nil), 400)
	if updated {
		t.Fatal("reset accepted bad code")
	}
	expect(t, call(t, a, "POST", "/api/reset-password", map[string]string{"email": "a@example.com", "token": "12345678", "password": "NewStrong123"}, nil), 200)
	if !updated {
		t.Fatal("password not updated")
	}
	var n int
	a.db.QueryRow("SELECT count(*) FROM sessions WHERE user_id=?", id).Scan(&n)
	if n != 0 {
		t.Fatal("old session remains")
	}
}
func TestSupabaseSignupNeverStoresPassword(t *testing.T) {
	a := sbSetup(t, func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Password string      `json:"password"`
			Data     credentials `json:"data"`
		}
		json.NewDecoder(r.Body).Decode(&in)
		if in.Password != "StrongPass123" || in.Data.Password != "" {
			t.Fatal("password metadata leak")
		}
		reply(w, 200, map[string]string{"id": "new-id"})
	})
	expect(t, call(t, a, "POST", "/api/register", credentials{Name: "Student", Email: "new@example.com", Password: "StrongPass123", Role: "student", Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}, nil), 201)
	var n int
	a.db.QueryRow("SELECT count(*) FROM users").Scan(&n)
	if n != 0 {
		t.Fatal("unverified signup created profile")
	}
}
func TestSupabaseDoesNotRelinkDifferentIdentity(t *testing.T) {
	a := sbSetup(t, func(w http.ResponseWriter, r *http.Request) {})
	s := sbUser{ID: "first", Email: "new@example.com", Confirmed: "yes", Metadata: credentials{Name: "New User", Role: "student", Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}}
	if _, err := a.sbProfile(s); err != nil {
		t.Fatal(err)
	}
	s.ID = "replacement"
	if _, err := a.sbProfile(s); err == nil {
		t.Fatal("identity relink allowed")
	}
}
