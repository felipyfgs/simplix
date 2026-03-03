package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/simplix/api/internal/domain"
)

type QuePasaService struct {
	client *http.Client
}

func NewQuePasaService() *QuePasaService {
	return &QuePasaService{client: &http.Client{}}
}

// ValidateConnection verifies that the bot token is valid by fetching bot info.
func (s *QuePasaService) ValidateConnection(settings domain.QuePasaSettings) error {
	url := fmt.Sprintf("%s/api/v3/%s/me", settings.BaseURL, settings.BotToken)
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("quepasa unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("invalid token: %s", string(body))
	}
	return nil
}

// SendTextMessage sends a plain text message via QuePasa.
// Returns the QuePasa message ID on success.
func (s *QuePasaService) SendTextMessage(settings domain.QuePasaSettings, toPhone, content string) (string, error) {
	url := fmt.Sprintf("%s/api/v3/%s/sendtext", settings.BaseURL, settings.BotToken)
	payload := map[string]string{
		"chatid": toPhone,
		"text":   content,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("quepasa error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err == nil && result.ID != "" {
		return result.ID, nil
	}
	return "", nil
}

// RegisterWebhook registers the inbound webhook URL with the QuePasa bot.
func (s *QuePasaService) RegisterWebhook(settings domain.QuePasaSettings, webhookURL string) error {
	url := fmt.Sprintf("%s/api/v3/%s/webhook", settings.BaseURL, settings.BotToken)
	payload := map[string]string{"url": webhookURL}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("quepasa unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("quepasa webhook registration failed: %s", string(body))
	}
	return nil
}
