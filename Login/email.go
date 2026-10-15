package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func (a *App) sendEmail(to, link string) error {
	if a.cfg.EmailMode == "console" {
		log.Printf("LOCAL DEVELOPMENT ONLY — verification for %s: %s", to, link)
		return nil
	}
	if a.cfg.EmailMode == "smtp" {
		return a.sendGmail(to, link)
	}
	body, e := json.Marshal(map[string]any{"from": a.cfg.EmailFrom, "to": []string{to}, "subject": "Verify your ResearchBridge account", "text": fmt.Sprintf("Welcome to ResearchBridge!\n\nVerify your email by opening this link, then pressing Verify email:\n%s\n\nThe link expires in 30 minutes. If you did not register, ignore this email.", link)})
	if e != nil {
		return e
	}
	req, e := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	response, e := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if e != nil {
		return errors.New("email service unavailable")
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("email service status %d", response.StatusCode)
	}
	return nil
}
