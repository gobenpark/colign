package email

import (
	"fmt"
	"log"
)

type Sender interface {
	SendVerificationEmail(to, token string) error
}

type LogSender struct {
	BaseURL string
}

func NewLogSender(baseURL string) *LogSender {
	return &LogSender{BaseURL: baseURL}
}

func (s *LogSender) SendVerificationEmail(to, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.BaseURL, token)
	log.Printf("[EMAIL] To: %s | Verify URL: %s", to, verifyURL)
	return nil
}
