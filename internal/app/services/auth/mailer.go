package auth

import (
	"fmt"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
)

// EmailSender sends emails. Implement this interface to send password reset links in production.
// ForgotPassword sends the raw reset token only through this interface.
type EmailSender interface {
	SendPasswordResetEmail(to, resetToken string) error
}

// ResolveMailer picks the mailer for the current environment.
func ResolveMailer() (EmailSender, error) {
	if config.IsDevelopment() {
		return NewLogMailer(), nil
	}
	driver := config.GetString("MAIL_DRIVER")
	if driver == "log" {
		return NewLogMailer(), nil
	}
	if driver == "smtp" {
		return NewSMTPMailer()
	}
	return nil, fmt.Errorf("MAIL_DRIVER must be set to smtp or log outside development (got %q)", driver)
}
