package telegram

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sender/internal/entities/notification"
	"strconv"
	"strings"

	"github.com/wb-go/wbf/zlog"
)

type SendMessage struct {
	TelegramID string `json:"chat_id"`
	Message    string `json:"text"`
}

func Send(n notification.Notification) {
	const op = "internal.service.telegram.send"
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage", os.Getenv("BOT_TOKEN"),
	)
	body := fmt.Sprintf(
		`{"chat_id": "%s", "text": "%s"}`,
		strconv.FormatInt(n.TelegramID, 10), n.Message,
	)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).Send()
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).Send()
		return
	}
	if resp.StatusCode != http.StatusOK {
		b := new(bytes.Buffer)
		_, _ = io.Copy(b, resp.Body)
		zlog.Logger.Error().Err(errors.New("wrong status code on request")).
			Fields(map[string]any{"op": op, "body": b.String()}).Send()
	}
}
