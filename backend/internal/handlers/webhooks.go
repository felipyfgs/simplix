package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/repository"
)

type WebhookHandler struct{ webhooks *repository.WebhookRepo }

func NewWebhookHandler(wh *repository.WebhookRepo) *WebhookHandler { return &WebhookHandler{webhooks: wh} }

func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	whs, err := h.webhooks.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, whs)
}

func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL           string   `json:"url"`
		Subscriptions []string `json:"subscriptions"`
		Secret        *string  `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		Error(w, http.StatusBadRequest, "url required")
		return
	}
	wh, err := h.webhooks.Create(r.Context(), req.URL, req.Subscriptions, req.Secret)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, wh)
}

func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		URL           string   `json:"url"`
		Subscriptions []string `json:"subscriptions"`
		Enabled       bool     `json:"enabled"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	wh, err := h.webhooks.Update(r.Context(), id, req.URL, req.Subscriptions, req.Enabled)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, wh)
}

func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	h.webhooks.Delete(r.Context(), id)
	w.WriteHeader(http.StatusNoContent)
}
