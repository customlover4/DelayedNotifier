package request

import (
	"delayednotifier/internal/entities/notification"
	"strconv"
	"strings"
	"time"
)

type UpdateNotification struct {
	Status string `json:"status"`
}

func (u *UpdateNotification) Validate() (status, msg string) {
	if u.Status == notification.StatusPending {
		return notification.StatusPending, ""
	}
	if u.Status == notification.StatusComplete {
		return notification.StatusComplete, ""
	}

	return "", "wrong status"
}

// CreateNotification модель запроса для создания нового уведомления
type CreateNotification struct {
	Message    string `json:"message"`
	TelegramID string `json:"telegram_id"`
	Email      string `json:"email"`
	Date       string `json:"date"`
}

func (c *CreateNotification) Validate() (notification.Notification, string) {
	r := notification.Notification{}
	if c.Message == "" {
		return notification.Notification{}, "message is empty"
	}
	r.Message = c.Message
	if c.TelegramID == "" && c.Email == "" {
		return notification.Notification{}, "both send variant is empty"
	}
	if c.TelegramID != "" {
		tID, err := strconv.ParseInt(c.TelegramID, 10, 64)
		if err != nil {
			return notification.Notification{}, "telegram_id, shoud be numeric"
		} else if tID <= 0 {
			return notification.Notification{}, "wrong telegram_id, shoud be >0"
		}
		r.TelegramID = tID
	}
	if c.Email != "" {
		if !strings.Contains(c.Email, "@") || !strings.Contains(c.Email, ".") {
			return notification.Notification{}, "wrong email format"
		}
	}
	r.Email = c.Email
	t, err := time.Parse(notification.DateLayout, c.Date)
	if err != nil {
		return notification.Notification{}, "wrong date value (format: 2006-01-02 15:04)"
	}
	r.Date = t

	return r, ""
}
