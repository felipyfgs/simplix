package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
	"github.com/simplix/api/internal/service"
)

type InboxHandler struct {
	inboxes     *repository.InboxRepo
	whatsappSvc *service.WhatsAppService
	quepasaSvc  *service.QuePasaService
	publicURL   string
}

func NewInboxHandler(inboxes *repository.InboxRepo, wa *service.WhatsAppService, qp *service.QuePasaService, publicURL string) *InboxHandler {
	return &InboxHandler{inboxes: inboxes, whatsappSvc: wa, quepasaSvc: qp, publicURL: publicURL}
}

func (h *InboxHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.inboxes.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []domain.Inbox{}
	}
	JSON(w, http.StatusOK, maskInboxes(list))
}

func (h *InboxHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	ix, err := h.inboxes.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "inbox not found")
		return
	}
	JSON(w, http.StatusOK, maskInbox(*ix))
}

func (h *InboxHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		ChannelType string `json:"channel_type"`
		// WhatsApp fields
		PhoneNumber       string `json:"phone_number"`
		PhoneNumberID     string `json:"phone_number_id"`
		BusinessAccountID string `json:"business_account_id"`
		APIKey            string `json:"api_key"`
		// QuePasa fields
		BotToken string `json:"bot_token"`
		BaseURL  string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		Error(w, http.StatusBadRequest, "name required")
		return
	}

	if req.ChannelType == string(domain.ChannelQuePasa) {
		h.createQuePasa(w, r, req.Name, req.BotToken, req.PhoneNumber, req.BaseURL)
		return
	}

	// Default: WhatsApp Cloud API
	if req.PhoneNumber == "" || req.PhoneNumberID == "" || req.BusinessAccountID == "" || req.APIKey == "" {
		Error(w, http.StatusBadRequest, "phone_number, phone_number_id, business_account_id and api_key required")
		return
	}

	settings := domain.WhatsAppSettings{
		PhoneNumber:        req.PhoneNumber,
		PhoneNumberID:      req.PhoneNumberID,
		BusinessAccountID:  req.BusinessAccountID,
		APIKey:             req.APIKey,
		WebhookVerifyToken: generateToken(),
	}

	if err := h.whatsappSvc.ValidateCredentials(settings); err != nil {
		Error(w, http.StatusBadRequest, "invalid WhatsApp credentials: "+err.Error())
		return
	}

	ix, err := h.inboxes.Create(r.Context(), req.Name, domain.ChannelWhatsApp, settings)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, maskInbox(*ix))
}

func (h *InboxHandler) createQuePasa(w http.ResponseWriter, r *http.Request, name, botToken, phoneNumber, baseURL string) {
	if botToken == "" || phoneNumber == "" {
		Error(w, http.StatusBadRequest, "bot_token and phone_number required")
		return
	}
	if baseURL == "" {
		baseURL = "http://quepasa:31000"
	}

	settings := domain.QuePasaSettings{
		BotToken:    botToken,
		PhoneNumber: phoneNumber,
		BaseURL:     baseURL,
	}

	if err := h.quepasaSvc.ValidateConnection(settings); err != nil {
		Error(w, http.StatusBadRequest, "invalid QuePasa credentials: "+err.Error())
		return
	}

	ix, err := h.inboxes.Create(r.Context(), name, domain.ChannelQuePasa, settings)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Auto-register webhook if PUBLIC_URL is configured
	if h.publicURL != "" {
		webhookURL := h.publicURL + "/webhook/quepasa/" + ix.ID.String()
		if err := h.quepasaSvc.RegisterWebhook(settings, webhookURL); err != nil {
			// Log but don't fail — user can register manually
		}
	}

	JSON(w, http.StatusCreated, maskInbox(*ix))
}

func (h *InboxHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	existing, err := h.inboxes.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "inbox not found")
		return
	}

	var req struct {
		Name string `json:"name"`
		// WhatsApp fields
		PhoneNumber       string `json:"phone_number"`
		PhoneNumberID     string `json:"phone_number_id"`
		BusinessAccountID string `json:"business_account_id"`
		APIKey            string `json:"api_key"`
		// QuePasa fields
		BotToken string `json:"bot_token"`
		BaseURL  string `json:"base_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}

	var settings any
	switch existing.ChannelType {
	case domain.ChannelQuePasa:
		current := quePasaSettingsFromInbox(existing)
		if req.BotToken != "" {
			current.BotToken = req.BotToken
		}
		if req.PhoneNumber != "" {
			current.PhoneNumber = req.PhoneNumber
		}
		if req.BaseURL != "" {
			current.BaseURL = req.BaseURL
		}
		settings = current
	default:
		current := whatsAppSettingsFromInbox(existing)
		if req.PhoneNumber != "" {
			current.PhoneNumber = req.PhoneNumber
		}
		if req.PhoneNumberID != "" {
			current.PhoneNumberID = req.PhoneNumberID
		}
		if req.BusinessAccountID != "" {
			current.BusinessAccountID = req.BusinessAccountID
		}
		if req.APIKey != "" && req.APIKey != "***" {
			current.APIKey = req.APIKey
		}
		settings = current
	}

	ix, err := h.inboxes.Update(r.Context(), id, name, settings)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, maskInbox(*ix))
}

func (h *InboxHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.inboxes.Delete(r.Context(), id); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *InboxHandler) SyncTemplates(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	ix, err := h.inboxes.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "inbox not found")
		return
	}
	if ix.ChannelType != domain.ChannelWhatsApp {
		Error(w, http.StatusBadRequest, "not a WhatsApp inbox")
		return
	}

	settings := whatsAppSettingsFromInbox(ix)
	templates, err := h.whatsappSvc.SyncTemplates(settings)
	if err != nil {
		Error(w, http.StatusBadGateway, "failed to sync templates: "+err.Error())
		return
	}

	settings.MessageTemplates = templates
	if err := h.inboxes.UpdateSettings(r.Context(), id, settings); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]any{"synced": len(templates), "templates": templates})
}

func (h *InboxHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	ix, err := h.inboxes.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "inbox not found")
		return
	}
	settings := whatsAppSettingsFromInbox(ix)
	templates := settings.MessageTemplates
	if templates == nil {
		templates = []any{}
	}
	JSON(w, http.StatusOK, templates)
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func whatsAppSettingsFromInbox(ix *domain.Inbox) domain.WhatsAppSettings {
	if s, ok := ix.Settings.(domain.WhatsAppSettings); ok {
		return s
	}
	return domain.WhatsAppSettings{}
}

func quePasaSettingsFromInbox(ix *domain.Inbox) domain.QuePasaSettings {
	if s, ok := ix.Settings.(domain.QuePasaSettings); ok {
		return s
	}
	return domain.QuePasaSettings{}
}

func maskInbox(ix domain.Inbox) map[string]any {
	out := map[string]any{
		"id":           ix.ID,
		"name":         ix.Name,
		"channel_type": ix.ChannelType,
		"created_at":   ix.CreatedAt,
	}
	switch s := ix.Settings.(type) {
	case domain.WhatsAppSettings:
		out["settings"] = map[string]any{
			"phone_number":          s.PhoneNumber,
			"phone_number_id":       s.PhoneNumberID,
			"business_account_id":   s.BusinessAccountID,
			"api_key":               "***",
			"webhook_verify_token":  s.WebhookVerifyToken,
			"templates_last_synced": s.TemplatesLastSynced,
		}
	case domain.QuePasaSettings:
		out["settings"] = map[string]any{
			"phone_number": s.PhoneNumber,
			"base_url":     s.BaseURL,
		}
	}
	return out
}

func maskInboxes(list []domain.Inbox) []map[string]any {
	out := make([]map[string]any, len(list))
	for i, ix := range list {
		out[i] = maskInbox(ix)
	}
	return out
}

func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
