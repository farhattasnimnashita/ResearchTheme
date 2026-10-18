package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

// Supabase passwords and refresh tokens are never stored in the app database.
// App sessions last at most one hour. Dashboard authorization still uses local roles.
type sbUser struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	Confirmed string      `json:"email_confirmed_at"`
	Metadata  credentials `json:"user_metadata"`
}
type sbResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	User        sbUser `json:"user"`
}

var sbHTTP = &http.Client{Timeout: 60* time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}

func initSupabaseTables(db *Database) error {
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS supabase_identities(auth_id TEXT PRIMARY KEY,user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE)`,
		`CREATE TABLE IF NOT EXISTS supabase_sessions(hash TEXT PRIMARY KEY REFERENCES sessions(hash) ON DELETE CASCADE)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}
func (a *App) sbCall(ctx context.Context, method, path, bearer string, input, output any) error {
	b, err := json.Marshal(input)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, a.cfg.SupabaseURL+"/auth/v1"+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("apikey", a.cfg.SupabaseKey)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	res, err := sbHTTP.Do(req)
	if err != nil {
    println("Supabase connection error:", err.Error())
    return errors.New("Authentication service unavailable. Try again shortly.")
}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if res.StatusCode == 429 {
			return errors.New("Too many attempts. Please wait before trying again.")
		}
		var problem struct {
			Code string `json:"error_code"`
		}
		json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&problem)
		println("Supabase error:", res.StatusCode, problem.Code)
		switch problem.Code {
		case "email_not_confirmed":
			return errors.New("Verify your email first. Open Verify email or resend the code.")
		case "user_already_exists", "email_exists":
			return errors.New("An account already exists. Log in or reset your password.")
		case "weak_password":
			return errors.New("Choose a stronger password that meets your Supabase password policy.")
		default:
			return errors.New("Authentication request failed. Check your details/code; for email delivery problems check Supabase Auth logs and SMTP settings.")
		}
	}
	if output != nil {
		if err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(output); err != nil {
			return errors.New("Invalid authentication service response.")
		}
	}
	return nil
}
func (a *App) sbRegister(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !decode(w, r, &c) {
		return
	}
	em, ok := email(c.Email)
	c.Name = strings.TrimSpace(c.Name)
	if !ok || !role(c.Role) || len(c.Name) < 2 || len(c.Name) > 100 || len(c.Password) < 8 || len(c.Password) > 72 {
		fail(w, 400, "Enter a valid name, email, role and password (8–72 bytes).")
		return
	}
	if err := validateProfile(&c); err != nil {
		fail(w, 400, err.Error())
		return
	}
	password := c.Password
	c.Password = ""
	c.Email = em
	// Local profiles are created only after Supabase proves email ownership.
	var result json.RawMessage
	if err := a.sbCall(r.Context(), "POST", "/signup", "", map[string]any{"email": em, "password": password, "data": c}, &result); err != nil {
		fail(w, 502, err.Error())
		return
	}
	reply(w, 201, map[string]string{"message": "Check your inbox and Spam for the verification code.", "redirect": "/verify"})
}
func (a *App) sbResend(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if !decode(w, r, &in) {
		return
	}
	em, ok := email(in.Email)
	if !ok {
		fail(w, 400, "Enter a valid email.")
		return
	}
	if err := a.sbCall(r.Context(), "POST", "/resend", "", map[string]string{"email": em, "type": "signup"}, nil); err != nil {
		fail(w, 502, err.Error())
		return
	}
	reply(w, 200, map[string]string{"message": "If confirmation is needed, a code will arrive. Open Verify email and enter it.", "redirect": "/verify"})
}
func (a *App) sbVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
		Token string `json:"token"`
	}
	if !decode(w, r, &in) {
		return
	}
	em, ok := email(in.Email)
	if !ok || !validOTP(in.Token) {
		fail(w, 400, "Enter your email and the code from your email.")
		return
	}
	var result sbResponse
	if err := a.sbCall(r.Context(), "POST", "/verify", "", map[string]string{"email": em, "token": in.Token, "type": "email"}, &result); err != nil {
		fail(w, 400, err.Error())
		return
	}
	if result.User.ID == "" || result.User.Confirmed == "" {
		fail(w, 400, "Email confirmation was not completed.")
		return
	}
	// Verification alone never grants an app session; users sign in with a password.
	reply(w, 200, map[string]string{"message": "Email verified. Return to login and sign in with your password."})
}
func validOTP(s string) bool {
	if len(s) < 6 || len(s) > 10 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Existing profiles are linked only after a successful password sign-in AND
// verified email ownership. An existing role/profile cannot be overwritten by metadata.
func (a *App) sbProfile(s sbUser) (User, error) {
	var u User
	em, ok := email(s.Email)
	if !ok || s.ID == "" || s.Confirmed == "" {
		return u, errors.New("Verify your email before signing in.")
	}
	tx, err := a.db.Begin()
	if err != nil {
		return u, err
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRow("SELECT user_id FROM supabase_identities WHERE auth_id=?", s.ID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow("SELECT id FROM users WHERE email=?", em).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			c := s.Metadata
			c.Name = strings.TrimSpace(c.Name)
			if !role(c.Role) || len(c.Name) < 2 || len(c.Name) > 100 {
				return u, errors.New("Missing profile. Register through ResearchTheme first.")
			}
			if err = validateProfile(&c); err != nil {
				return u, err
			}
			id, err = tx.InsertID(`INSERT INTO users(name,email,password_hash,role,verified,created_at,institution,department,academic_year,semester) VALUES(?,?,?,?,?,?,?,?,?,?)`, c.Name, em, "!supabase-auth", c.Role, 1, time.Now().Unix(), c.Institution, c.Department, c.AcademicYear, c.Semester)
		}
		if err != nil {
			return u, err
		}
		if _, err = tx.Exec("INSERT INTO supabase_identities(auth_id,user_id) VALUES(?,?)", s.ID, id); err != nil {
			return u, errors.New("Account already linked to another identity. Contact the project owner.")
		}
		// Disable the old local password and invalidate pre-migration sessions.
		if _, err = tx.Exec("DELETE FROM sessions WHERE user_id=?", id); err != nil {
			return u, err
		}
	}
	if err != nil {
		return u, err
	}
	// Email is immutable in this app; dashboard changes are not silently imported.
	if err = tx.QueryRow("SELECT id,name,email,role FROM users WHERE id=?", id).Scan(&u.ID, &u.Name, &u.Email, &u.Role); err != nil {
		return u, err
	}
	if u.Email != em {
		return u, errors.New("Account email changed. Contact the project owner to reconcile your profile.")
	}
	if _, err = tx.Exec("UPDATE users SET verified=1,password_hash=? WHERE id=?", "!supabase-auth", id); err != nil {
		return u, err
	}
	u.Verified = true
	return u, tx.Commit()
}
func (a *App) sbLogin(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if !decode(w, r, &c) {
		return
	}
	em, ok := email(c.Email)
	if !ok || !role(c.Role) || len(c.Password) < 8 || len(c.Password) > 72 {
		fail(w, 400, "Enter valid login details.")
		return
	}
	var result sbResponse
	if err := a.sbCall(r.Context(), "POST", "/token?grant_type=password", "", map[string]string{"email": em, "password": c.Password}, &result); err != nil {
		fail(w, 401, err.Error())
		return
	}
	if result.AccessToken == "" || result.User.Confirmed == "" {
		fail(w, 403, "Email verification is required.")
		return
	}
	u, err := a.sbProfile(result.User)
	if err != nil {
		fail(w, 409, "Could not link your profile. If you changed the account email or recreated the Auth user, contact the project owner.")
		return
	}
	if u.Role != c.Role {
		fail(w, 403, "Choose the role used for your existing account.")
		return
	}
	t, err := token()
	if err != nil {
		fail(w, 500, "Could not create session.")
		return
	}
	ttl := result.ExpiresIn
	if ttl <= 0 || ttl > 3600 {
		ttl = 3600
	}
	tx, err := a.db.Begin()
	if err != nil {
		fail(w, 500, "Could not create session.")
		return
	}
	defer tx.Rollback()
	if _, err = tx.Exec("INSERT INTO sessions(hash,user_id,expires) VALUES(?,?,?)", digest(t), u.ID, time.Now().Add(time.Duration(ttl)*time.Second).Unix()); err == nil {
		_, err = tx.Exec("INSERT INTO supabase_sessions(hash) VALUES(?)", digest(t))
	}
	if old, e := r.Cookie("rb_session"); e == nil && err == nil {
		_, err = tx.Exec("DELETE FROM sessions WHERE hash=?", digest(old.Value))
	}
	if err != nil || tx.Commit() != nil {
		fail(w, 500, "Could not create session.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "rb_session", Value: t, Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.cfg.URL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: ttl})
	reply(w, 200, map[string]string{"redirect": "/dashboard/" + u.Role})
}
func (a *App) sbForgot(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if !decode(w, r, &in) {
		return
	}
	em, ok := email(in.Email)
	if !ok {
		fail(w, 400, "Enter a valid email.")
		return
	}
	if err := a.sbCall(r.Context(), "POST", "/recover", "", map[string]string{"email": em}, nil); err != nil {
		fail(w, 502, err.Error())
		return
	}
	reply(w, 200, map[string]string{"message": "If an Auth account exists, a recovery code will arrive. Check inbox and Spam.", "redirect": "/reset-password"})
}
func (a *App) sbReset(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email"`
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	em, ok := email(in.Email)
	if !ok || !validOTP(in.Token) || len(in.Password) < 8 || len(in.Password) > 72 {
		fail(w, 400, "Enter your email, recovery code and new password (8–72 bytes).")
		return
	}
	var result sbResponse
	if err := a.sbCall(r.Context(), "POST", "/verify", "", map[string]string{"email": em, "token": in.Token, "type": "recovery"}, &result); err != nil {
		fail(w, 400, err.Error())
		return
	}
	if result.AccessToken == "" || result.User.ID == "" || result.User.Confirmed == "" || !strings.EqualFold(result.User.Email, em) {
		fail(w, 400, "Invalid recovery response.")
		return
	}
	// Revoke existing app sessions before changing the password. An unmigrated
	// profile is matched by a now-verified email, without creating a new profile.
	if _, err := a.db.Exec("DELETE FROM sessions WHERE user_id IN (SELECT user_id FROM supabase_identities WHERE auth_id=?) OR user_id IN (SELECT id FROM users WHERE email=?)", result.User.ID, em); err != nil {
		fail(w, 500, "Could not revoke old sessions. Request a new recovery code.")
		return
	}
	if err := a.sbCall(r.Context(), "PUT", "/user", result.AccessToken, map[string]string{"password": in.Password}, nil); err != nil {
		fail(w, 400, err.Error()+" Request a fresh recovery code before retrying.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "rb_session", Path: "/", HttpOnly: true, Secure: strings.HasPrefix(a.cfg.URL, "https://"), SameSite: http.SameSiteLaxMode, MaxAge: -1})
	reply(w, 200, map[string]string{"message": "Password updated. Sign in with your new password."})
}
