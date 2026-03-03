package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/simplix/api/internal/domain"
)

const metaAPIBase = "https://graph.facebook.com/v18.0"

type WhatsAppService struct {
	client *http.Client
}

func NewWhatsAppService() *WhatsAppService {
	return &WhatsAppService{client: &http.Client{}}
}

type TemplateComponent struct {
	Type       string               `json:"type"`
	SubType    string               `json:"sub_type,omitempty"`
	Index      *int                 `json:"index,omitempty"`
	Parameters []TemplateParameter  `json:"parameters,omitempty"`
}

type TemplateParameter struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
}

type TemplatePayload struct {
	Name       string              `json:"name"`
	LangCode   string              `json:"lang_code"`
	Components []TemplateComponent `json:"components,omitempty"`
}

// ValidateCredentials verifies the API key and business account ID are valid.
func (s *WhatsAppService) ValidateCredentials(settings domain.WhatsAppSettings) error {
	url := fmt.Sprintf("%s/%s/message_templates?access_token=%s&limit=1",
		metaAPIBase, settings.BusinessAccountID, settings.APIKey)
	resp, err := s.client.Get(url)
	if err != nil {
		return fmt.Errorf("meta API unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("invalid credentials: %s", string(body))
	}
	return nil
}

// SendTextMessage sends a plain text message via WhatsApp Cloud API.
// Returns the Meta message ID on success.
func (s *WhatsAppService) SendTextMessage(settings domain.WhatsAppSettings, toPhone, content string) (string, error) {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toPhone,
		"type":              "text",
		"text":              map[string]string{"body": content},
	}
	return s.postMessage(settings, payload)
}

// SendTemplate sends a WhatsApp template message.
func (s *WhatsAppService) SendTemplate(settings domain.WhatsAppSettings, toPhone string, tpl TemplatePayload) (string, error) {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                toPhone,
		"type":              "template",
		"template": map[string]any{
			"name":       tpl.Name,
			"language":   map[string]string{"code": tpl.LangCode},
			"components": tpl.Components,
		},
	}
	return s.postMessage(settings, payload)
}

// SyncTemplates fetches all approved templates from Meta and returns them.
func (s *WhatsAppService) SyncTemplates(settings domain.WhatsAppSettings) ([]any, error) {
	return s.fetchTemplates(settings, fmt.Sprintf("%s/%s/message_templates?access_token=%s&limit=100",
		metaAPIBase, settings.BusinessAccountID, settings.APIKey))
}

func (s *WhatsAppService) fetchTemplates(settings domain.WhatsAppSettings, url string) ([]any, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meta API error: %d", resp.StatusCode)
	}

	var result struct {
		Data   []any `json:"data"`
		Paging *struct {
			Next string `json:"next"`
		} `json:"paging"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	templates := result.Data
	if result.Paging != nil && result.Paging.Next != "" {
		next, err := s.fetchTemplates(settings, result.Paging.Next)
		if err == nil {
			templates = append(templates, next...)
		}
	}
	return templates, nil
}

// VerifyToken checks that the provided token matches the stored webhook verify token.
func (s *WhatsAppService) VerifyToken(settings domain.WhatsAppSettings, token string) bool {
	return settings.WebhookVerifyToken != "" && settings.WebhookVerifyToken == token
}

func (s *WhatsAppService) postMessage(settings domain.WhatsAppSettings, payload map[string]any) (string, error) {
	url := fmt.Sprintf("%s/%s/messages", metaAPIBase, settings.PhoneNumberID)

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("meta API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	if len(result.Messages) > 0 {
		return result.Messages[0].ID, nil
	}
	return "", nil
}
