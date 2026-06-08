// Package notify provides outbound notification clients.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const telegramAPI = "https://api.telegram.org/bot%s/sendMessage"
const telegramPhotoAPI = "https://api.telegram.org/bot%s/sendPhoto"

// Telegram sends messages through the Telegram Bot API.
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
		ChatID    string `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode,omitempty"`
	}{
		ChatID:    t.ChatID,
		Text:      message,
		ParseMode: "HTML",
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

// SendPhoto posts a PNG image with an optional HTML caption.
func (t Telegram) SendPhoto(filename string, image []byte, caption string) error {
	if t.Token == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if t.ChatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", t.ChatID); err != nil {
		return fmt.Errorf("write chat id: %w", err)
	}
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return fmt.Errorf("write caption: %w", err)
		}
		if err := writer.WriteField("parse_mode", "HTML"); err != nil {
			return fmt.Errorf("write parse mode: %w", err)
		}
	}

	part, err := writer.CreateFormFile("photo", filename)
	if err != nil {
		return fmt.Errorf("create photo part: %w", err)
	}
	if _, err := part.Write(image); err != nil {
		return fmt.Errorf("write photo: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(telegramPhotoAPI, t.Token), &body)
	if err != nil {
		return fmt.Errorf("build telegram photo request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram photo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram photo API returned %s: %s", resp.Status, string(respBody))
	}

	return nil
}
