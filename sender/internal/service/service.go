package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sender/internal/entities/notification"
	"sender/internal/service/email"
	"sender/internal/service/telegram"
	"sender/internal/storage"
	"strings"

	"github.com/wb-go/wbf/zlog"
)

type Service struct {
	str   *storage.Storage
	email string
}

func New(str *storage.Storage, email string) *Service {
	return &Service{
		str:   str,
		email: email,
	}
}

var (
	ErrWrongStatusCode = errors.New("status code of request not 200")
)

func UpdateStatus(id int64) error {
	port := os.Getenv("NOTIFIER_PORT")
	body := fmt.Sprintf(`{"status": "%s"}`, notification.StatusComplete)
	req, err := http.NewRequest(http.MethodPatch,
		fmt.Sprintf("http://notifier:%s/notify/%d", port, id),
		strings.NewReader(body),
	)
	if err != nil {
		return err
	}
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		b := new(bytes.Buffer)
		_, _ = io.Copy(b, resp.Body)
		zlog.Logger.Info().Fields(map[string]any{"body": b.String()}).
			Send()
		return ErrWrongStatusCode
	}

	return nil
}

func (s *Service) Start() {
	const op = "internal.servce.MainCycle"
	for msg := range s.str.Receiver() {
		n := notification.Notification{}
		err := n.UnmarshalBinary(msg.Body)
		if err != nil {
			zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).
				Send()
			continue
		}
		err = UpdateStatus(n.ID)
		if err != nil {
			zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).
				Send()
			continue
		}
		if n.TelegramID != 0 {
			go telegram.Send(n)
		}
		if n.Email != "" {
			go email.Send(n, s.email)
		}
	}
}
