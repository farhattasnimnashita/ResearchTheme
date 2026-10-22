package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

//go:embed web
var frontend embed.FS

type Config struct {
	Addr, URL, DBPath, Env, EmailMode, APIKey, EmailFrom string
	SMTPUser, SMTPPass                                   string
	DatabaseURL                                          string
	SupabaseURL, SupabaseKey                             string
	StorageMode, StorageKey, StorageBucket               string
}
type App struct {
	db        *Database
	cfg       Config
	send      func(string, string) error
	mu        sync.Mutex
	limits    map[string]window
	resetOnce sync.Once
	resetJobs chan string
}
type window struct {
	start time.Time
	count int
}
type User struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	Verified     bool   `json:"verified"`
	Institution  string `json:"institution"`
	Department   string `json:"department"`
	AcademicYear int    `json:"academic_year"`
	Semester     int    `json:"semester"`
}
type credentials struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	Institution  string `json:"institution"`
	Department   string `json:"department"`
	AcademicYear int    `json:"academic_year"`
	Semester     int    `json:"semester"`
}

func loadEnv() error {
	f, err := os.Open(".env")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			return errors.New("invalid .env line")
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), "\"'")
		if _, set := os.LookupEnv(k); !set {
			if err = os.Setenv(k, v); err != nil {
				return err
			}
		}
	}
	return s.Err()
}
func env(k, d string) string {
	if s := os.Getenv(k); s != "" {
		return s
	}
	return d
}
func config() (Config, error) {
	c := Config{
		DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")),
		SupabaseURL: strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/"), SupabaseKey: strings.TrimSpace(os.Getenv("SUPABASE_PUBLISHABLE_KEY")),
		Addr: env("APP_ADDR", "127.0.0.1:8080"), URL: env("APP_URL", "http://localhost:8080"),
		DBPath: env("DB_PATH", "researchbridge.db"), Env: env("APP_ENV", "development"),
		EmailMode: env("EMAIL_MODE", "console"), APIKey: os.Getenv("RESEND_API_KEY"),
		EmailFrom: os.Getenv("EMAIL_FROM"), SMTPUser: strings.TrimSpace(os.Getenv("SMTP_USER")),
		SMTPPass:      strings.ReplaceAll(strings.TrimSpace(os.Getenv("SMTP_PASS")), " ", ""),
		StorageMode:   env("STORAGE_MODE", "local"),
		StorageKey:    strings.TrimSpace(os.Getenv("STORAGE_KEY")),
		StorageBucket: env("STORAGE_BUCKET", "research-files"),
	}
	u, e := url.Parse(c.URL)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return c, errors.New("APP_URL must be an http(s) origin without trailing slash")
	}
	if c.EmailMode != "console" && c.EmailMode != "resend" && c.EmailMode != "smtp" && c.EmailMode != "supabase" {
		return c, errors.New("EMAIL_MODE must be console, resend, smtp or supabase")
	}
	if c.Env != "development" && c.Env != "production" {
		return c, errors.New("APP_ENV must be development or production")
	}
	if c.Env == "production" && (u.Scheme != "https" || c.EmailMode == "console") {
		return c, errors.New("production requires HTTPS APP_URL and a real email mode (resend or smtp)")
	}
	if c.EmailMode == "console" {
		host, _, err := net.SplitHostPort(c.Addr)
		ip := net.ParseIP(host)
		if err != nil || ip == nil || !ip.IsLoopback() || !(u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1" || u.Hostname() == "::1") {
			return c, errors.New("console email is only allowed on a loopback address and localhost APP_URL")
		}
	}
	if c.EmailMode == "resend" {
		if c.APIKey == "" {
			return c, errors.New("RESEND_API_KEY is required")
		}
		if _, err := mail.ParseAddress(c.EmailFrom); err != nil {
			return c, errors.New("EMAIL_FROM must be a valid sender")
		}
	}
	if c.EmailMode == "smtp" {
		em, ok := email(c.SMTPUser)
		if !ok || em != c.SMTPUser || !strings.HasSuffix(em, "@gmail.com") {
			return c, errors.New("SMTP_USER must be your full personal Gmail address (lowercase)")
		}
		if len(c.SMTPPass) != 16 {
			return c, errors.New("SMTP_PASS must be the 16-character Google App Password, not your normal Gmail password")
		}
		if c.EmailFrom == "" {
			c.EmailFrom = "ResearchBridge <" + c.SMTPUser + ">"
		}
		from, err := mail.ParseAddress(c.EmailFrom)
		if err != nil || strings.ContainsAny(c.EmailFrom, "\r\n") || !strings.EqualFold(from.Address, c.SMTPUser) {
			return c, errors.New("EMAIL_FROM must use the same Gmail address as SMTP_USER")
		}
	}
	if c.EmailMode == "supabase" {
		endpoint, err := url.Parse(c.SupabaseURL)
		if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" || endpoint.Path != "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" || c.SupabaseKey == "" {
			return c, errors.New("SUPABASE_URL must be an HTTPS origin and SUPABASE_PUBLISHABLE_KEY is required")
		}
		if strings.HasPrefix(c.SupabaseKey, "sb_secret_") {
			return c, errors.New("Use the publishable key, not a Supabase secret key")
		}
	}
	return c, nil
}
func openSQLite(path string) (*sql.DB, error) {
	db, e := sql.Open("sqlite", path)
	if e != nil {
		return nil, e
	}
	db.SetMaxOpenConns(1)
	_, e = db.Exec(`PRAGMA foreign_keys=ON; PRAGMA busy_timeout=5000; PRAGMA journal_mode=WAL;
 CREATE TABLE IF NOT EXISTS users(id INTEGER PRIMARY KEY, name TEXT NOT NULL,email TEXT NOT NULL UNIQUE,password_hash TEXT NOT NULL,role TEXT NOT NULL CHECK(role IN ('student','teacher')),verified INTEGER NOT NULL DEFAULT 0,created_at INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS verifications(hash TEXT PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,expires INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS sessions(hash TEXT PRIMARY KEY,user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,expires INTEGER NOT NULL);
 CREATE TABLE IF NOT EXISTS projects(id INTEGER PRIMARY KEY,student_id INTEGER NOT NULL REFERENCES users(id),teacher_id INTEGER REFERENCES users(id),title TEXT NOT NULL,area TEXT NOT NULL,abstract TEXT NOT NULL,status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','submitted','approved','changes_requested','rejected')),feedback TEXT NOT NULL DEFAULT '',updated_at INTEGER NOT NULL);`)
	if e != nil {
		db.Close()
		return nil, e
	}
	if e = migrateProjectProposalColumns(db); e != nil {
		db.Close()
		return nil, e
	}
	if e = migrateUserProfile(db); e != nil {
		db.Close()
		return nil, e
	}
	if e = migrateWorkspace(db); e != nil {
		db.Close()
		return nil, e
	}

	if e = migrateProfileExtras(&Database{DB: db}); e != nil {
		db.Close()
		return nil, e
	}

	return db, nil
}
func token() (string, error) {
	b := make([]byte, 32)
	_, e := rand.Read(b)
	return hex.EncodeToString(b), e
}
func digest(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
func fail(w http.ResponseWriter, status int, s string) {
	reply(w, status, map[string]string{"error": s})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil {
		fail(w, 400, "Invalid request.")
		return false
	}
	if d.Decode(&struct{}{}) != io.EOF {
		fail(w, 400, "Only one JSON object is allowed.")
		return false
	}
	return true
}
func email(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	a, e := mail.ParseAddress(s)
	return s, e == nil && a.Address == s && len(s) <= 254
}

func role(s string) bool { return s == "student" || s == "teacher" }
func (a *App) routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /api/auth-config", func(w http.ResponseWriter, r *http.Request) {
		reply(w, 200, map[string]bool{"supabase": a.cfg.EmailMode == "supabase"})
	})
	m.HandleFunc("POST /api/student-profile", a.editStudentProfile)
	m.HandleFunc("GET /api/profile-picture", a.getProfilePicture)
	m.HandleFunc("POST /api/profile-picture", a.uploadProfilePicture)
	m.HandleFunc("GET /api/students", a.searchStudents)
	m.HandleFunc("POST /api/team/invite", a.inviteDirectoryStudent)
	
	m.HandleFunc("POST /api/forgot-password", func(w http.ResponseWriter, r *http.Request) {
		if a.cfg.EmailMode == "supabase" {
			a.sbForgot(w, r)
			return
		}

		a.forgotPassword(w, r)
	})

	m.HandleFunc("POST /api/reset-password", func(w http.ResponseWriter, r *http.Request) {
		if a.cfg.EmailMode == "supabase" {
			a.sbReset(w, r)
			return
		}

		a.resetPassword(w, r)
	})
	m.HandleFunc("GET /api/researcher-profile", a.researcherProfile)
	m.HandleFunc("POST /api/researcher-profile", a.researcherProfile)
	m.HandleFunc("GET /api/collaboration-projects", a.collaborationProjects)
	m.HandleFunc("GET /api/workspace/{id}/members", a.workspaceMembers)
	m.HandleFunc("POST /api/workspace/{id}/members", a.workspaceMembers)
	m.HandleFunc("GET /api/invitations", a.invitations)
	m.HandleFunc("POST /api/invitations", a.invitations)
	m.HandleFunc("GET /api/workspace/{id}/records", a.records)
	m.HandleFunc("POST /api/workspace/{id}/records", a.records)
	m.HandleFunc("GET /api/records/{id}/history", a.recordHistory)
	m.HandleFunc("GET /api/workspace/{id}/permissions", a.permissions)
	m.HandleFunc("POST /api/workspace/{id}/permissions", a.permissions)
	m.HandleFunc("GET /api/workspace/{id}/files", a.workspaceFiles)
	m.HandleFunc("POST /api/workspace/{id}/files", a.workspaceFiles)
	m.HandleFunc("GET /api/files/{id}", a.downloadFile)
	m.HandleFunc("GET /api/notifications", a.notifications)
	m.HandleFunc("POST /api/notifications", a.notifications)
	m.HandleFunc("POST /api/register", a.register)
	m.HandleFunc("POST /api/login", a.login)
	m.HandleFunc("POST /api/verify", a.verify)
	m.HandleFunc("POST /api/resend", a.resend)
	m.HandleFunc("POST /api/logout", a.logout)
	m.HandleFunc("GET /api/me", func(w http.ResponseWriter, r *http.Request) {
		u, e := a.user(r)
		if e != nil {
			fail(w, 401, "Please log in.")
			return
		}
		reply(w, 200, u)
	})
	m.HandleFunc("GET /api/teachers", a.teachers)
	m.HandleFunc("GET /api/teacher-directory", a.teacherDirectory)
	m.HandleFunc("GET /api/projects", a.listProjects)
	m.HandleFunc("POST /api/projects", a.createProject)
	m.HandleFunc(
	"POST /api/projects/{id}/supervisor",
	a.selectProjectSupervisor,
)
	m.HandleFunc("POST /api/projects/{id}/submit", a.submitProject)
	m.HandleFunc("POST /api/projects/{id}/review", a.reviewProject)
	m.HandleFunc(
		"POST /api/projects/{id}/proposal",
		a.saveProjectProposal,
	)

	m.HandleFunc(
		"POST /api/projects/{id}/proposal/submit",
		a.submitDetailedProjectProposal,
	)
	assets, _ := fs.Sub(frontend, "web/assets")
	m.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	m.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p != "/" && p != "/login/student" && p != "/login/teacher" && p != "/signup" && p != "/verify" && p != "/dashboard/student" && p != "/dashboard/teacher" && p != "/workspace/student" && p != "/research" && p != "/forgot-password" && p != "/reset-password" {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(p, "/dashboard/") || p == "/workspace/student" || p == "/research" {
			u, e := a.user(r)
			if e != nil {
				role := strings.TrimPrefix(p, "/dashboard/")
				if p == "/workspace/student" || p == "/research" {
					role = "student"
				}
				http.Redirect(w, r, "/login/"+role, 303)
				return
			}
			if p != "/dashboard/"+u.Role && !(p == "/workspace/student" && u.Role == "student") && p != "/research" {
				http.Redirect(w, r, "/dashboard/"+u.Role, 303)
				return
			}
		}
		page := "web/index.html"
		if p == "/dashboard/student" {
			page = "web/student-dashboard.html"
		}
		if p == "/dashboard/teacher" || p == "/research" {
			page = "web/teacher-dashboard.html"
		}
		if p == "/forgot-password" || p == "/reset-password" {
			page = "web/password-recovery.html"
		}
		if a.cfg.EmailMode == "supabase" && (p == "/verify" || p == "/forgot-password" || p == "/reset-password") {
			page = "web/supabase-otp.html"
		}
		b, e := frontend.ReadFile(page)
		if e != nil {
			http.Error(w, "Page unavailable", 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(b)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self'; script-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method == "POST" {
			if r.Header.Get("Origin") != a.cfg.URL ||
				r.Header.Get("X-Requested-With") != "ResearchBridge" ||
				!(strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") ||
					(strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data;") &&
						(strings.HasSuffix(r.URL.Path, "/files") ||
							r.URL.Path == "/api/profile-picture"))) {
				fail(w, 403, "Invalid request origin. Open the app using APP_URL.")
				return
			}
		}
		if r.Method == "POST" && r.URL.Path != "/api/logout" && !a.allowed(r) {
			w.Header().Set("Retry-After", "60")
			fail(w, 429, "Too many attempts. Please try again in one minute.")
			return
		}
		m.ServeHTTP(w, r)
	})
}
func (a *App) allowed(r *http.Request) bool {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.limits == nil {
		a.limits = map[string]window{}
	}
	now := time.Now()
	for k, v := range a.limits {
		if now.Sub(v.start) > time.Minute {
			delete(a.limits, k)
		}
	}
	v := a.limits[host]
	if v.start.IsZero() {
		v.start = now
	}
	if v.count >= 30 {
		return false
	}
	v.count++
	a.limits[host] = v
	return true
}
func (a *App) user(r *http.Request) (User, error) {
	var u User
	c, e := r.Cookie("rb_session")
	if e != nil {
		return u, e
	}
	if a.cfg.EmailMode == "supabase" {
		var present int
		if a.db.QueryRow("SELECT 1 FROM supabase_sessions WHERE hash=?", digest(c.Value)).Scan(&present) != nil {
			return u, errors.New("sign in again")
		}
	}
	e = a.db.QueryRow(`SELECT u.id,u.name,u.email,u.role,u.verified,u.institution,u.department,u.academic_year,u.semester FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.hash=? AND s.expires>? AND u.verified=1`, digest(c.Value), time.Now().Unix()).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Verified, &u.Institution, &u.Department, &u.AcademicYear, &u.Semester)
	return u, e
}
func (a *App) register(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbRegister(w, r)
		return
	}
	var c credentials
	if !decode(w, r, &c) {
		return
	}
	c.Name = strings.TrimSpace(c.Name)
	em, ok := email(c.Email)
	if !ok || !role(c.Role) || len(c.Name) < 2 || len(c.Name) > 100 || len(c.Password) < 8 || len(c.Password) > 72 {
		fail(w, 400, "Enter a name (2–100 characters), valid email, role and password (8–72 bytes).")
		return
	}
	if err := validateProfile(&c); err != nil {
		fail(w, 400, err.Error())
		return
	}
	h, e := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if e != nil {
		fail(w, 500, "Could not secure password.")
		return
	}
	id, e := a.db.InsertID(`INSERT INTO users(name,email,password_hash,role,created_at,institution,department,academic_year,semester) VALUES(?,?,?,?,?,?,?,?,?)`, c.Name, em, string(h), c.Role, time.Now().Unix(), c.Institution, c.Department, c.AcademicYear, c.Semester)
	if e != nil {
		var id int64
		if a.db.QueryRow("SELECT id FROM users WHERE email=?", em).Scan(&id) == nil {
			fail(w, 409, "This email already has an account. Log in or resend verification.")
		} else {
			fail(w, 500, "Could not save account.")
		}
		return
	}
	if e = a.issueVerification(id, em); e != nil {
		log.Printf("Verification delivery failed for account %d", id)
		reply(w, 201, map[string]any{"message": "Account saved, but verification could not be sent. Use Resend verification after checking the email configuration.", "delivery_failed": true})
		return
	}
	reply(w, 201, map[string]string{"message": a.verificationMessage()})
}
func (a *App) verificationMessage() string {
	if a.cfg.EmailMode == "console" {
		return "Account saved. Development mode: open the verification link printed in the server terminal. No email was sent."
	}
	return "Check your inbox and spam folder for your verification link."
}
func (a *App) issueVerification(id int64, em string) error {
	t, e := token()
	if e != nil {
		return e
	}
	_, e = a.db.Exec("INSERT INTO verifications(hash,user_id,expires) VALUES(?,?,?)", digest(t), id, time.Now().Add(30*time.Minute).Unix())
	if e != nil {
		return e
	}
	link := a.cfg.URL + "/verify?token=" + url.QueryEscape(t)
	if e = a.send(em, link); e != nil {
		a.db.Exec("DELETE FROM verifications WHERE hash=?", digest(t))
		return e
	}
	return nil
}
func (a *App) resend(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbResend(w, r)
		return
	}
	var c struct {
		Email string `json:"email"`
	}
	if !decode(w, r, &c) {
		return
	}
	em, ok := email(c.Email)
	if !ok {
		fail(w, 400, "Enter a valid email.")
		return
	}
	var id int64
	var verified bool
	if a.db.QueryRow("SELECT id,verified FROM users WHERE email=?", em).Scan(&id, &verified) == nil && !verified {
		if e := a.issueVerification(id, em); e != nil {
			fail(w, 502, "Could not send verification. Check email configuration and try again.")
			return
		}
	}
	msg := "If an unverified account exists, a new verification email has been sent. Check your inbox and spam folder."
	if a.cfg.EmailMode == "console" {
		msg = "If an unverified account exists, a new verification link is in the server terminal. Development console mode does not send email."
	}
	reply(w, 200, map[string]string{"message": msg})
}
func (a *App) verify(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbVerify(w, r)
		return
	}
	var c struct {
		Token string `json:"token"`
	}
	if !decode(w, r, &c) {
		return
	}
	if len(c.Token) != 64 {
		fail(w, 400, "Invalid verification link.")
		return
	}
	tx, e := a.db.Begin()
	if e != nil {
		fail(w, 500, "Could not verify account.")
		return
	}
	defer tx.Rollback()
	var id int64
	e = tx.QueryRow("SELECT user_id FROM verifications WHERE hash=? AND expires>?", digest(c.Token), time.Now().Unix()).Scan(&id)
	if e != nil {
		fail(w, 400, "Verification link expired or already used. Request a new link.")
		return
	}
	var userRole string
	e = tx.QueryRow("SELECT role FROM users WHERE id=?", id).Scan(&userRole)
	if e == nil {
		_, e = tx.Exec("UPDATE users SET verified=1 WHERE id=?", id)
	}
	if e == nil {
		_, e = tx.Exec("DELETE FROM verifications WHERE user_id=?", id)
	}
	if e == nil {
		e = tx.Commit()
	}
	if e != nil {
		fail(w, 500, "Could not verify account.")
		return
	}
	reply(w, 200, map[string]string{"message": "Email verified. You can now log in.", "role": userRole})
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("unused-timing-password"), bcrypt.DefaultCost)

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbLogin(w, r)
		return
	}
	var c credentials
	if !decode(w, r, &c) {
		return
	}
	em, ok := email(c.Email)
	if !ok || !role(c.Role) || len(c.Password) > 72 {
		fail(w, 400, "Enter valid login details.")
		return
	}
	var u User
	var h string
	e := a.db.QueryRow("SELECT id,name,email,role,verified,password_hash FROM users WHERE email=?", em).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.Verified, &h)
	if errors.Is(e, sql.ErrNoRows) {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(c.Password))
		reply(w, 404, map[string]any{"error": "No account found with this email. Please sign up to create your account.", "code": "account_not_found", "suggest_signup": true})
		return
	}
	if e != nil {
		fail(w, 500, "Could not check your account. Please try again.")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(h), []byte(c.Password)) != nil {
		reply(w, 401, map[string]string{"error": "Incorrect password. Use the password you created when signing up on ResearchBridge.", "code": "invalid_password"})
		return
	}
	if u.Role != c.Role {
		reply(w, 403, map[string]string{"error": "This account uses the other role. Select the role you chose during signup.", "code": "role_mismatch"})
		return
	}
	if !u.Verified {
		fail(w, 403, "Verify your email before logging in. Use Resend verification if needed.")
		return
	}
	t, e := token()
	if e != nil {
		fail(w, 500, "Could not create session.")
		return
	}
	a.db.Exec("DELETE FROM sessions WHERE expires<=?", time.Now().Unix())
	_, e = a.db.Exec("INSERT INTO sessions(hash,user_id,expires) VALUES(?,?,?)", digest(t), u.ID, time.Now().Add(24*time.Hour).Unix())
	if e != nil {
		fail(w, 500, "Could not save session.")
		return
	}
	if old, err := r.Cookie("rb_session"); err == nil {
		a.db.Exec("DELETE FROM sessions WHERE hash=?", digest(old.Value))
	}
	http.SetCookie(w, &http.Cookie{Name: "rb_session", Value: t, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.cfg.URL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	reply(w, 200, map[string]string{"redirect": "/dashboard/" + u.Role})
}
func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	if c, e := r.Cookie("rb_session"); e == nil {
		if _, e = a.db.Exec("DELETE FROM sessions WHERE hash=?", digest(c.Value)); e != nil {
			fail(w, 500, "Could not log out. Please retry.")
			return
		}
	}
	http.SetCookie(w, &http.Cookie{Name: "rb_session", Path: "/", MaxAge: -1, HttpOnly: true, Secure: strings.HasPrefix(a.cfg.URL, "https://"), SameSite: http.SameSiteLaxMode})
	reply(w, 200, map[string]bool{"ok": true})
}
func main() {
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) && os.Getenv("EMAIL_MODE") == "" {
		folder, _ := os.Getwd()
		log.Fatalf("No .env in %s. Copy your existing Gmail-configured .env into this folder before running. Email mode will not silently fall back to console.", folder)
	}
	if e := loadEnv(); e != nil {
		log.Fatal(e)
	}
	c, e := config()
	if e != nil {
		log.Fatal(e)
	}
	db, e := openConfiguredDB(c)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	if c.EmailMode == "supabase" {
		if e = initSupabaseTables(db); e != nil {
			log.Fatal("Could not initialize Supabase Auth mappings: ", e)
		}
	}
	fmt.Printf("Database mode: %s\n", db.Mode())
	a := &App{db: db, cfg: c}
	a.send = a.sendEmail
	server := &http.Server{Addr: c.Addr, Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 90 * time.Second, IdleTimeout: 60 * time.Second}
	fmt.Printf("ResearchBridge: %s\nEmail mode: %s\n", c.URL, c.EmailMode)
	log.Fatal(server.ListenAndServe())
}
