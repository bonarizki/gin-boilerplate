package auth

import (
	"fmt"

	"github.com/bonarizki/gin-boilerplate/pkg/logger"
)

// LogMailer logs password reset tokens instead of sending email.
// Intended for local development only.
type LogMailer struct{}

// NewLogMailer returns a development mailer that writes reset links to logs.
func NewLogMailer() *LogMailer {
	return &LogMailer{}
}

func (m *LogMailer) SendPasswordResetEmail(to, resetToken string) error {
	logger.Warnf("DEV password reset for %s — token: %s", to, resetToken)
	fmt.Printf("[DEV] password reset for %s — token: %s\n", to, resetToken)
	return nil
}