package service

import (
	"delayednotifier/internal/entities/notification"
	"delayednotifier/internal/storage"
	"errors"
	"fmt"
	"strings"
	"time"
)

type storager interface {
	CreateNotification(n notification.Notification) (int64, error)
	GetNotification(id int64) (notification.Notification, error)
	DeleteNotification(id int64) error
	UpdateNotificationStatus(status string, id int64) error
}

type Service struct {
	str storager
}

func New(s storager) *Service {
	return &Service{
		str: s,
	}
}

var (
	ErrNotValidData    = errors.New("not valid data")
	ErrStorageInternal = errors.New("internal error in storage")
	ErrNotFound        = errors.New("not found")
	ErrNotAffected     = errors.New("no one didn't be affected")
)

func (s *Service) CreateNotification(n notification.Notification) (int64, error) {
	if n.Date.Unix() < time.Now().UTC().Add(time.Second*20).Unix() {
		return 0, fmt.Errorf(
			"%w: %s", ErrNotValidData, "date in past",
		)
	}
	if n.TelegramID != 0 {
		if n.TelegramID <= 0 {
			return 0, fmt.Errorf(
				"%w: %s", ErrNotValidData, "telegram_id can't be <= 0",
			)
		}
	}
	if n.Email != "" {
		if !strings.Contains(n.Email, "@") || !strings.Contains(n.Email, ".") {
			return 0, fmt.Errorf(
				"%w: %s", ErrNotValidData, "not valid email format",
			)
		}

	}

	id, err := s.str.CreateNotification(n)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrStorageInternal, err)
	}

	return id, nil
}

func (s *Service) Notification(id int64) (notification.Notification, error) {
	if id <= 0 {
		return notification.Notification{}, fmt.Errorf(
			"%w: %s", ErrNotValidData, "notification id is negative or == 0",
		)
	}

	n, err := s.str.GetNotification(id)
	if errors.Is(err, storage.ErrNotFound) {
		return n, fmt.Errorf("%w: %w", ErrNotFound, err)
	} else if err != nil {
		return n, fmt.Errorf("%w: %w", ErrStorageInternal, err)
	}

	return n, nil
}

func (s *Service) DeleteNotification(id int64) error {
	if id <= 0 {
		return fmt.Errorf(
			"%w: %s", ErrNotValidData, "notification id is negative or == 0",
		)
	}

	err := s.str.DeleteNotification(id)
	if errors.Is(err, storage.ErrNotAffected) {
		return fmt.Errorf("%w: %w", ErrNotAffected, err)
	} else if err != nil {
		return fmt.Errorf("%w: %w", ErrStorageInternal, err)
	}

	return nil
}

func (s *Service) UpdateNotificationStatus(status string, id int64) error {
	if status != notification.StatusComplete && status != notification.StatusPending {
		return fmt.Errorf(
			"%w: %s", ErrNotValidData,
			"wrong status value (\"pending\" or \"complete\" only)",
		)
	}
	if id <= 0 {
		return fmt.Errorf(
			"%w: %s", ErrNotValidData, "notification id is negative or == 0",
		)
	}

	err := s.str.UpdateNotificationStatus(status, id)
	if errors.Is(err, storage.ErrNotAffected) {
		return ErrNotAffected
	} else if err != nil {
		return fmt.Errorf("%w: %w", ErrStorageInternal, err)
	}

	return nil
}
