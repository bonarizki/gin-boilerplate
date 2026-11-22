package auth

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
)

// SMTPMailer sends password reset emails via SMTP.
type SMTPMailer struct {
	host     string
	port     string
	user     string
	password string
	from     string
	baseURL  string
}

// NewSMTPMailer builds an SMTP mailer from configuration.
func NewSMTPMailer() (*SMTPMailer, error) {
	host := config.GetString("SMTP_HOST")
	port := config.GetString("SMTP_PORT")
	from := config.GetString("MAIL_FROM")
	if host == "" || port == "" || from == "" {
		return nil, fmt.Errorf("SMTP_HOST, SMTP_PORT, and MAIL_FROM are required when MAIL_DRIVER=smtp")
	}
	return &SMTPMailer{
		host:     host,
		port:     port,
		user:     config.GetString("SMTP_USER"),
		password: config.GetString("SMTP_PASS"),
		from:     from,
		baseURL:  strings.TrimRight(config.GetString("RESET_URL_BASE"), "/"),
	}, nil
}

func (m *SMTPMailer) SendPasswordResetEmail(to, resetToken string) error {
	link := resetToken
	if m.baseURL != "" {
		link = fmt.Sprintf("%s?token=%s", m.baseURL, resetToken)
	}
	body := fmt.Sprintf("To reset your password, open: %s\n\nIf you did not request this, ignore this message.", link)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: Password reset\r\n\r\n%s", to, body))
	addr := m.host + ":" + m.port
	var auth smtp.Auth
	if m.user != "" {
		auth = smtp.PlainAuth("", m.user, m.password, m.host)
	}
	return smtp.SendMail(addr, auth, m.from, []string{to}, msg)
}