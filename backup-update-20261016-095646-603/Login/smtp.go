package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// Gmail uses STARTTLS on port 587. TLS is mandatory before authentication.
// This sender is deliberately limited to personal Gmail accounts for this starter.
func (a *App) sendGmail(to, link string) error {
	message, err := verificationMail(a.cfg.EmailFrom, to, link)
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{Timeout: 10 * time.Second}).Dial("tcp", "smtp.gmail.com:587")
	if err != nil {
		return errors.New("cannot connect to Gmail SMTP; check your network/firewall")
	}
	defer conn.Close()
	if err = conn.SetDeadline(time.Now().Add(20 * time.Second)); err != nil {
		return err
	}
	err = smtpExchange(conn, "smtp.gmail.com", &tls.Config{ServerName: "smtp.gmail.com", MinVersion: tls.VersionTLS12}, a.cfg.SMTPUser, a.cfg.SMTPPass, to, message)
	if err != nil {
		// Provider replies may contain account details; never log passwords or message bodies.
		return errors.New("Gmail SMTP delivery failed; check App Password, account and network")
	}
	return nil
}

func smtpExchange(conn net.Conn, host string, config *tls.Config, user, password, to string, message []byte) error {
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return errors.New("SMTP server does not offer mandatory STARTTLS")
	}
	if err = c.StartTLS(config); err != nil {
		return err
	}
	if err = c.Auth(smtp.PlainAuth("", user, password, host)); err != nil {
		return err
	}
	if err = c.Mail(user); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}
	writer, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(message); err != nil {
		writer.Close()
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	// The server has now accepted DATA. A QUIT failure must not undo a delivered token.
	_ = c.Quit()
	return nil
}

func verificationMail(from, to, link string) ([]byte, error) {
	if strings.ContainsAny(from+to+link, "\r\n") {
		return nil, errors.New("invalid email header or link")
	}
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return nil, err
	}
	recipient, err := mail.ParseAddress(to)
	if err != nil || recipient.Address != to {
		return nil, errors.New("invalid recipient")
	}
	id, err := token()
	if err != nil {
		return nil, err
	}
	subject, message := accountMailText(link)
	var out bytes.Buffer
	fmt.Fprintf(&out, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s@%s>\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", sender.String(), recipient.String(), subject, time.Now().Format(time.RFC1123Z), id, strings.SplitN(sender.Address, "@", 2)[1])
	body := quotedprintable.NewWriter(&out)
	_, err = body.Write([]byte(message))
	if err != nil {
		return nil, err
	}
	if err = body.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
