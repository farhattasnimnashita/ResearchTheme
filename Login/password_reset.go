package main

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Link tokens are hashed at rest and consumed atomically; recovery revokes every session.
func (a *App) forgotPassword(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbForgot(w, r)
		return
	}
	var in struct {
		Email string `json:"email"`
	}
	if !decode(w, r, &in) {
		return
	}
	em, ok := email(in.Email)
	if !ok {
		fail(w, 400, "Enter a valid email address.")
		return
	}
	now := time.Now().Unix()
	// A per-address cooldown applies even to unknown accounts and across source IPs.
	tx, err := a.db.Begin()
	if err != nil {
		fail(w, 500, "Could not request recovery.")
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM reset_requests WHERE requested_at<?", now-3600)
	if err != nil {
		fail(w, 500, "Could not request recovery.")
		return
	}
	var recent int64
	err = tx.QueryRow("SELECT requested_at FROM reset_requests WHERE email_hash=?", digest(em)).Scan(&recent)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		fail(w, 500, "Could not request recovery.")
		return
	}
	if recent > now-60 {
		reply(w, 200, map[string]string{"message": "If this account exists, a reset link will arrive by email. Please wait a minute before requesting another."})
		return
	}
	_, err = tx.Exec("INSERT INTO reset_requests(email_hash,requested_at) VALUES(?,?) ON CONFLICT(email_hash) DO UPDATE SET requested_at=excluded.requested_at", digest(em), now)
	if err != nil || tx.Commit() != nil {
		fail(w, 500, "Could not request recovery.")
		return
	}
	// Queue delivery for both known and unknown addresses; response timing never waits for SMTP.
	a.resetOnce.Do(func() {
		a.resetJobs = make(chan string, 32)
		go func() {
			for address := range a.resetJobs {
				a.deliverReset(address)
			}
		}()
	})
	select {
	case a.resetJobs <- em:
	default:
		log.Printf("Password recovery queue is full; retry later")
	}

	reply(w, 200, map[string]string{"message": "If this account exists, a reset link will arrive by email. Check Inbox and Spam. The link expires in 30 minutes."})
}
func (a *App) resetPassword(w http.ResponseWriter, r *http.Request) {
	if a.cfg.EmailMode == "supabase" {
		a.sbReset(w, r)
		return
	}
	var in struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !decode(w, r, &in) {
		return
	}
	if len(in.Token) != 64 || len(in.Password) < 8 || len(in.Password) > 72 || strings.TrimSpace(in.Password) == "" {
		fail(w, 400, "Use a valid reset link and a password of 8–72 bytes.")
		return
	}
	h, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		fail(w, 500, "Could not update password.")
		return
	}
	tx, err := a.db.Begin()
	if err != nil {
		fail(w, 500, "Could not update password.")
		return
	}
	defer tx.Rollback()
	var id int64
	if tx.QueryRow("SELECT user_id FROM password_resets WHERE hash=? AND expires>?", digest(in.Token), time.Now().Unix()).Scan(&id) != nil {
		fail(w, 400, "Reset link expired or already used. Request a new link.")
		return
	}
	if _, err = tx.Exec("UPDATE users SET password_hash=? WHERE id=?", string(h), id); err == nil {
		_, err = tx.Exec("DELETE FROM password_resets WHERE user_id=?", id)
	}
	if err == nil {
		_, err = tx.Exec("DELETE FROM sessions WHERE user_id=?", id)
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		fail(w, 500, "Could not update password.")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "rb_session", Path: "/", MaxAge: -1, HttpOnly: true, Secure: strings.HasPrefix(a.cfg.URL, "https://"), SameSite: http.SameSiteLaxMode})
	reply(w, 200, map[string]string{"message": "Password updated. All previous sessions are signed out. Log in with your new password."})
}

func (a *App) deliverReset(em string) {
	var id int64
	if a.db.QueryRow("SELECT id FROM users WHERE email=?", em).Scan(&id) != nil {
		return
	}
	t, err := token()
	if err != nil {
		return
	}
	if _, err = a.db.Exec("INSERT INTO password_resets(hash,user_id,expires) VALUES(?,?,?)", digest(t), id, time.Now().Add(30*time.Minute).Unix()); err != nil {
		return
	}
	if err = a.send(em, a.cfg.URL+"/reset-password?token="+t); err != nil {
		a.db.Exec("DELETE FROM password_resets WHERE hash=?", digest(t))
		log.Printf("Password reset delivery failed; check configured email service")
	}
}
