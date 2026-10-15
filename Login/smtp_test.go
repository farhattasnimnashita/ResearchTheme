package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGmailConfiguration(t *testing.T) {
	for _, k := range []string{"APP_ADDR", "APP_URL", "DB_PATH", "APP_ENV", "EMAIL_MODE", "RESEND_API_KEY", "EMAIL_FROM", "SMTP_USER", "SMTP_PASS"} {
		t.Setenv(k, "")
	}
	t.Setenv("EMAIL_MODE", "smtp")
	if _, e := config(); e == nil {
		t.Fatal("missing Gmail credentials accepted")
	}
	t.Setenv("SMTP_USER", "sender@gmail.com")
	t.Setenv("SMTP_PASS", "abcd efgh ijkl mnop")
	c, e := config()
	if e != nil {
		t.Fatal(e)
	}
	if c.SMTPPass != "abcdefghijklmnop" || c.EmailFrom != "ResearchBridge <sender@gmail.com>" {
		t.Fatal("incorrect credential normalization or default sender")
	}
	t.Setenv("EMAIL_FROM", "other@gmail.com")
	if _, e = config(); e == nil {
		t.Fatal("different sender accepted")
	}
	t.Setenv("EMAIL_FROM", "")
	t.Setenv("SMTP_PASS", "regular-password")
	// The length alone cannot prove a Google-issued secret; no authentication happens until sending.
	t.Setenv("SMTP_PASS", "short")
	if _, e = config(); e == nil {
		t.Fatal("short secret accepted")
	}
	t.Setenv("SMTP_PASS", "abcdefghijklmnop")
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_URL", "https://research.example.com")
	if _, e = config(); e != nil {
		t.Fatal("production Gmail SMTP rejected", e)
	}
}
func TestSMTPMessageEncoding(t *testing.T) {
	b, e := verificationMail("ResearchBridge <sender@gmail.com>", "recipient@example.com", "http://localhost:8080/verify?token=abcdef")
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(b, []byte("Subject: Verify your ResearchBridge account\r\n")) || !bytes.Contains(b, []byte("token=3Dabcdef")) {
		t.Fatal("missing MIME headers/encoded link")
	}
	if _, e = verificationMail("sender@gmail.com\r\nBcc: attacker@example.com", "recipient@example.com", "http://localhost"); e == nil {
		t.Fatal("header injection allowed")
	}
}
func TestSMTPRefusesCleartextAuth(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	client.SetDeadline(time.Now().Add(3 * time.Second))
	done := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(server)
		if _, e := io.WriteString(server, "220 localhost ESMTP\r\n"); e != nil {
			done <- e
			return
		}
		_, e := reader.ReadString('\n')
		if e == nil {
			_, e = io.WriteString(server, "250 localhost\r\n")
		}
		done <- e
	}()
	err := smtpExchange(client, "localhost", &tls.Config{ServerName: "localhost", MinVersion: tls.VersionTLS12}, "sender@gmail.com", "fake-test-secret", "recipient@example.com", []byte("test"))
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("cleartext was not refused: %v", err)
	}
	if e := <-done; e != nil {
		t.Fatal(e)
	}
}
func TestSMTPStartTLSDelivery(t *testing.T) {
	certServer := httptest.NewTLSServer(nil)
	defer certServer.Close()
	roots := x509.NewCertPool()
	roots.AddCert(certServer.Certificate())
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	client.SetDeadline(time.Now().Add(5 * time.Second))
	server.SetDeadline(time.Now().Add(5 * time.Second))
	done := make(chan error, 1)
	go func() {
		defer server.Close()
		current := server
		reader := bufio.NewReader(current)
		send := func(s string) error { _, e := io.WriteString(current, s); return e }
		if e := send("220 localhost ESMTP\r\n"); e != nil {
			done <- e
			return
		}
		for {
			line, e := reader.ReadString('\n')
			if e != nil {
				done <- e
				return
			}
			cmd := strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				if _, ok := current.(*tls.Conn); ok {
					e = send("250-localhost\r\n250 AUTH PLAIN\r\n")
				} else {
					e = send("250-localhost\r\n250 STARTTLS\r\n")
				}
			case cmd == "STARTTLS":
				e = send("220 ready\r\n")
				if e == nil {
					secured := tls.Server(current, certServer.TLS)
					e = secured.Handshake()
					current = secured
					reader = bufio.NewReader(current)
				}
			case strings.HasPrefix(cmd, "AUTH PLAIN"):
				e = send("235 authenticated\r\n")
			case strings.HasPrefix(cmd, "MAIL FROM:"), strings.HasPrefix(cmd, "RCPT TO:"):
				e = send("250 OK\r\n")
			case cmd == "DATA":
				e = send("354 send message\r\n")
				if e == nil {
					for {
						var part string
						part, e = reader.ReadString('\n')
						if e != nil || part == ".\r\n" {
							break
						}
					}
				}
				if e == nil {
					e = send("250 accepted\r\n")
				}
			case cmd == "QUIT":
				e = send("221 bye\r\n")
				done <- e
				return
			default:
				done <- io.ErrUnexpectedEOF
				return
			}
			if e != nil {
				done <- e
				return
			}
		}
	}()
	msg, e := verificationMail("ResearchBridge <sender@gmail.com>", "recipient@example.com", "http://localhost/verify?token=test")
	if e != nil {
		t.Fatal(e)
	}
	e = smtpExchange(client, "127.0.0.1", &tls.Config{ServerName: "127.0.0.1", RootCAs: roots, MinVersion: tls.VersionTLS12}, "sender@gmail.com", "fake-test-secret", "recipient@example.com", msg)
	if e != nil {
		t.Fatal(e)
	}
	if e = <-done; e != nil {
		t.Fatal(e)
	}
}
