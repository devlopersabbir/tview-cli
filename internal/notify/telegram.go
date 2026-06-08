// Package notify provides outbound notification clients.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const telegramAPI = "https://api.telegram.org/bot%s/sendMessage"

// Telegram sends plain text messages through the Telegram Bot API.
type Telegram struct {
	Token  string
	ChatID string
}

// Send posts message to the configured Telegram chat.
func (t Telegram) Send(message string) error {
	if t.Token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if t.ChatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	payload := struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: t.ChatID,
		Text:   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode telegram payload: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf(telegramAPI, t.Token),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned %s: %s", resp.Status, string(respBody))
	}

	return nil
}
