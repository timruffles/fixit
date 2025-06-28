package auth

import (
	"context"
	"log/slog"

	"github.com/aarondl/authboss/v3"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Mailer struct {
	apiKey   string
	fromName string
	fromAddr string
}

func NewMailer(apiKey, fromName, fromAddr string) *Mailer {
	return &Mailer{
		apiKey:   apiKey,
		fromName: fromName,
		fromAddr: fromAddr,
	}
}

func (m *Mailer) Send(ctx context.Context, email authboss.Email) error {
	from := mail.NewEmail(m.fromName, m.fromAddr)
	to := mail.NewEmail("", email.To[0])

	message := mail.NewSingleEmail(from, email.Subject, to, email.TextBody, email.HTMLBody)

	//client := sendgrid.NewSendClient(m.apiKey)
	//_, err := client.Send(message)
	slog.Info("stubbed email", "message", message)
	return nil
}
