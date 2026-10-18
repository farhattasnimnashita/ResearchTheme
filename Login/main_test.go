package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) (*App, map[string]string) {
	t.Helper()
	var db *Database
	var e error
	if connection := os.Getenv("POSTGRES_TEST_URL"); connection != "" {
		schema := fmt.Sprintf("rb_test_%d", time.Now().UnixNano())
		db, e = openPostgresSchema(connection, schema)
		if e == nil {
			t.Cleanup(func() { db.Exec("DROP SCHEMA " + schema + " CASCADE"); db.Close() })
		}
	} else {
		db, e = openDB(filepath.Join(t.TempDir(), "test.db"))
	}
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if !db.postgres {
			db.Close()
		}
	})
	links := map[string]string{}
	a := &App{db: db, cfg: Config{URL: "http://localhost:8080", EmailMode: "console"}}
	a.send = func(email, link string) error { links[email] = link; return nil }
	return a, links
}
func call(t *testing.T, a *App, method, path string, data any, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(data)
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.RemoteAddr = "127.0.0.1:12345"
	if method == "POST" {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", a.cfg.URL)
		r.Header.Set("X-Requested-With", "ResearchBridge")
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	return w
}
func expect(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	if w.Code != status {
		t.Fatalf("wanted %d, got %d: %s", status, w.Code, w.Body.String())
	}
}
func register(t *testing.T, a *App, em, role string) {
	t.Helper()
	expect(t, call(t, a, "POST", "/api/register", credentials{Name: "Test Researcher", Email: em, Password: "StrongPass123", Role: role, Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}, nil), 201)
}
func verifyUser(t *testing.T, a *App, links map[string]string, em string) string {
	t.Helper()
	u, e := url.Parse(links[em])
	if e != nil {
		t.Fatal(e)
	}
	token := u.Query().Get("token")
	expect(t, call(t, a, "POST", "/api/verify", map[string]string{"token": token}, nil), 200)
	return token
}
func loginUser(t *testing.T, a *App, em, role string) *http.Cookie {
	t.Helper()
	w := call(t, a, "POST", "/api/login", credentials{Email: em, Password: "StrongPass123", Role: role}, nil)
	expect(t, w, 200)
	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatal("session cookie missing")
	}
	c := cookies[0]
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode {
		t.Fatal("unsafe cookie")
	}
	return c
}
func makeProject(t *testing.T, a *App, c *http.Cookie) int64 {
	t.Helper()
	w := call(t, a, "POST", "/api/projects", map[string]any{"title": "Bengali research proposal", "area": "NLP", "abstract": "A detailed proposal about Bengali misinformation detection.", "teacher_id": 0}, c)
	expect(t, w, 201)
	var result struct {
		ID int64 `json:"id"`
	}
	if e := json.Unmarshal(w.Body.Bytes(), &result); e != nil {
		t.Fatal(e)
	}
	return result.ID
}
func TestAuthenticationLifecycle(t *testing.T) {
	a, links := setup(t)
	em := "student@example.com"
	register(t, a, em, "student")
	expect(t, call(t, a, "POST", "/api/register", credentials{Name: "Duplicate", Email: em, Password: "StrongPass123", Role: "student", Institution: "SUST", Department: "CSE", AcademicYear: 2, Semester: 4}, nil), 409)
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: em, Password: "StrongPass123", Role: "student"}, nil), 403)
	old := verifyUser(t, a, links, em)
	expect(t, call(t, a, "POST", "/api/verify", map[string]string{"token": old}, nil), 400)
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: em, Password: "StrongPass123", Role: "teacher"}, nil), 403)
	expect(t, call(t, a, "POST", "/api/login", credentials{Email: em, Password: "wrong", Role: "student"}, nil), 401)
	c := loginUser(t, a, em, "student")
	expect(t, call(t, a, "GET", "/api/me", nil, c), 200)
	w := call(t, a, "GET", "/dashboard/teacher", nil, c)
	expect(t, w, 303)
	if w.Header().Get("Location") != "/dashboard/student" {
		t.Fatal("incorrect role redirect")
	}
	var password, hash string
	if e := a.db.QueryRow("SELECT password_hash FROM users WHERE email=?", em).Scan(&password); e != nil {
		t.Fatal(e)
	}
	if password == "StrongPass123" || !strings.HasPrefix(password, "$2") {
		t.Fatal("password not hashed")
	}
	if e := a.db.QueryRow("SELECT hash FROM sessions").Scan(&hash); e != nil {
		t.Fatal(e)
	}
	if hash == c.Value {
		t.Fatal("session token stored in plaintext")
	}
	expect(t, call(t, a, "POST", "/api/logout", map[string]any{}, c), 200)
	expect(t, call(t, a, "GET", "/api/me", nil, c), 401)
	c = loginUser(t, a, em, "student")
	a.db.Exec("UPDATE sessions SET expires=?", time.Now().Add(-time.Minute).Unix())
	expect(t, call(t, a, "GET", "/api/me", nil, c), 401)
}
func TestVerificationExpiryAndDeliveryRecovery(t *testing.T) {
	a, links := setup(t)
	a.send = func(string, string) error { return errors.New("provider unavailable") }
	register(t, a, "recover@example.com", "student")
	var n int
	a.db.QueryRow("SELECT count(*) FROM verifications").Scan(&n)
	if n != 0 {
		t.Fatal("failed delivery left live token")
	}
	a.send = func(em, link string) error { links[em] = link; return nil }
	expect(t, call(t, a, "POST", "/api/resend", map[string]string{"email": "recover@example.com"}, nil), 200)
	u, _ := url.Parse(links["recover@example.com"])
	tok := u.Query().Get("token")
	a.db.Exec("UPDATE verifications SET expires=?", time.Now().Add(-time.Minute).Unix())
	expect(t, call(t, a, "POST", "/api/verify", map[string]string{"token": tok}, nil), 400)
	expect(t, call(t, a, "POST", "/api/resend", map[string]string{"email": "recover@example.com"}, nil), 200)
	verifyUser(t, a, links, "recover@example.com")
}
func TestProjectAccessAndReview(t *testing.T) {
	a, links := setup(t)
	accounts := []struct{ em, role string }{{"s@example.com", "student"}, {"other@example.com", "student"}, {"t@example.com", "teacher"}, {"otherteacher@example.com", "teacher"}}
	cookies := map[string]*http.Cookie{}
	ids := map[string]int64{}
	for _, acc := range accounts {
		register(t, a, acc.em, acc.role)
		verifyUser(t, a, links, acc.em)
		cookies[acc.em] = loginUser(t, a, acc.em, acc.role)
		var idsValue int64
		if e := a.db.QueryRow("SELECT id FROM users WHERE email=?", acc.em).Scan(&idsValue); e != nil {
			t.Fatal(e)
		}
		ids[acc.em] = idsValue
	}
	s := cookies["s@example.com"]
	teacher := cookies["t@example.com"]
	other := cookies["other@example.com"]
	otherTeacher := cookies["otherteacher@example.com"]
	id := makeProject(t, a, s)
	path := "/api/projects/" + strconvID(id)
	expect(t, call(t, a, "GET", "/api/projects", nil, nil), 401)
	expect(t, call(t, a, "POST", "/api/projects", map[string]any{}, teacher), 403)
	for _, c := range []*http.Cookie{teacher, other, otherTeacher} {
		w := call(t, a, "GET", "/api/projects", nil, c)
		expect(t, w, 200)
		if strings.TrimSpace(w.Body.String()) != "[]" {
			t.Fatal("draft or another student's proposal leaked")
		}
	}
	submission := map[string]any{"teacher_id": ids["t@example.com"], "abstract": "Updated Bengali research proposal with specific research methodology."}
	expect(t, call(t, a, "POST", path+"/submit", submission, other), 409)
	expect(t, call(t, a, "POST", path+"/submit", submission, s), 200)
	review := map[string]string{"status": "changes_requested", "feedback": "Please describe the dataset and evaluation method."}
	expect(t, call(t, a, "POST", path+"/review", review, s), 403)
	expect(t, call(t, a, "POST", path+"/review", review, otherTeacher), 409)
	expect(t, call(t, a, "POST", path+"/review", review, teacher), 200)
	expect(t, call(t, a, "POST", path+"/submit", submission, s), 200)
	review["status"] = "approved"
	expect(t, call(t, a, "POST", path+"/review", review, teacher), 200)
	// Reset only the rate-limit window; authorization assertions below must not be masked by it.
	a.mu.Lock()
	a.limits = nil
	a.mu.Unlock()
	expect(t, call(t, a, "POST", path+"/submit", submission, s), 409)
	expect(t, call(t, a, "POST", path+"/review", review, teacher), 409)
	w := call(t, a, "GET", "/api/projects", nil, s)
	expect(t, w, 200)
	if !strings.Contains(w.Body.String(), "approved") || !strings.Contains(w.Body.String(), review["feedback"]) {
		t.Fatal("review not visible to student")
	}
	w = call(t, a, "GET", "/api/projects", nil, otherTeacher)
	expect(t, w, 200)
	if strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatal("proposal leaked to another teacher")
	}
}
func strconvID(id int64) string { b, _ := json.Marshal(id); return string(b) }
func TestOriginValidationAndRateLimit(t *testing.T) {
	a, _ := setup(t)
	r := httptest.NewRequest("POST", "/api/logout", strings.NewReader(`{}`))
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Requested-With", "ResearchBridge")
	w := httptest.NewRecorder()
	a.routes().ServeHTTP(w, r)
	expect(t, w, 403)
	for i := 0; i < 30; i++ {
		expect(t, call(t, a, "POST", "/api/resend", map[string]string{"email": "missing@example.com"}, nil), 200)
	}
	expect(t, call(t, a, "POST", "/api/resend", map[string]string{"email": "missing@example.com"}, nil), 429)
}
func TestDatabasePersistence(t *testing.T) {
	p := filepath.Join(t.TempDir(), "persistent.db")
	db, e := openDB(p)
	if e != nil {
		t.Fatal(e)
	}
	_, e = db.Exec("INSERT INTO users(name,email,password_hash,role,created_at) VALUES('Persisted','saved@example.com','hash','student',1)")
	if e != nil {
		t.Fatal(e)
	}
	db.Close()
	db, e = openDB(p)
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	var name string
	if e = db.QueryRow("SELECT name FROM users WHERE email='saved@example.com'").Scan(&name); e != nil || name != "Persisted" {
		t.Fatal("record did not persist", e)
	}
}
func TestInvalidRoleAndExpiredSession(t *testing.T) {
	a, _ := setup(t)
	expect(t, call(t, a, "POST", "/api/register", credentials{Name: "Admin", Email: "admin@example.com", Password: "StrongPass123", Role: "admin"}, nil), 400)
	expect(t, call(t, a, "GET", "/dashboard/student", nil, nil), 303)
	expect(t, call(t, a, "GET", "/api/me", nil, &http.Cookie{Name: "rb_session", Value: "invalid"}), 401)
}

func TestProductionConfiguration(t *testing.T) {
	for _, key := range []string{"APP_ADDR", "APP_URL", "DB_PATH", "APP_ENV", "EMAIL_MODE", "RESEND_API_KEY", "EMAIL_FROM"} {
		t.Setenv(key, "")
	}
	if _, e := config(); e != nil {
		t.Fatal("safe local defaults failed", e)
	}
	t.Setenv("APP_ADDR", "0.0.0.0:8080")
	if _, e := config(); e == nil {
		t.Fatal("console mode must refuse public binding")
	}
	t.Setenv("APP_ENV", "production")
	t.Setenv("EMAIL_MODE", "resend")
	t.Setenv("APP_URL", "https://research.example.com")
	t.Setenv("RESEND_API_KEY", "fake-test-key")
	t.Setenv("EMAIL_FROM", "ResearchBridge <verify@example.com>")
	if _, e := config(); e != nil {
		t.Fatal("valid production config failed", e)
	}
	t.Setenv("APP_URL", "http://research.example.com")
	if _, e := config(); e == nil {
		t.Fatal("production allowed insecure HTTP")
	}
}
func TestFrontendAndVerificationGET(t *testing.T) {
	a, links := setup(t)
	register(t, a, "get@example.com", "student")
	u, _ := url.Parse(links["get@example.com"])
	for _, path := range []string{"/", "/login/student", "/login/teacher", "/signup", "/assets/app.js", "/assets/style.css", "/assets/library.png", u.RequestURI()} {
		w := call(t, a, "GET", path, nil, nil)
		expect(t, w, 200)
		if w.Header().Get("Content-Security-Policy") == "" {
			t.Fatal("missing CSP")
		}
	}
	var verified bool
	if e := a.db.QueryRow("SELECT verified FROM users WHERE email='get@example.com'").Scan(&verified); e != nil {
		t.Fatal(e)
	}
	if verified {
		t.Fatal("a link scanner GET must not verify the account")
	}
	verifyUser(t, a, links, "get@example.com")
	expect(t, call(t, a, "GET", "/.env", nil, nil), 404)
	expect(t, call(t, a, "GET", "/main.go", nil, nil), 404)
}
