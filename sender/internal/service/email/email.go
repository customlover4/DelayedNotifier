package email

import (
	"os"
	"sender/internal/entities/notification"

	"github.com/wb-go/wbf/zlog"
	gomail "gopkg.in/mail.v2"
)

func Send(n notification.Notification, emailUsername string) {
	const op = "internal.service.email.Send"

	m := gomail.NewMessage()
	m.SetHeader("From", emailUsername)
	m.SetHeader("To", n.Email)
	m.SetHeader("Subject", "Новое уведомление")
	m.SetBody("text/plain", n.Message)

	d := gomail.NewDialer(
		"smtp.gmail.com", 587, emailUsername,
		os.Getenv("EMAIL_PASSWORD"),
	)

	if err := d.DialAndSend(m); err != nil {
		zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).Send()
	}
}
