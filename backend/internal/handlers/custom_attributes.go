package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
)

type CustomAttributeHandler struct {
	attrs *repository.CustomAttributeRepo
}

func NewCustomAttributeHandler(attrs *repository.CustomAttributeRepo) *CustomAttributeHandler {
	return &CustomAttributeHandler{attrs: attrs}
}

func (h *CustomAttributeHandler) List(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	attrs, err := h.attrs.List(r.Context(), entityType)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if attrs == nil {
		attrs = []domain.CustomAttributeDefinition{}
	}
	JSON(w, http.StatusOK, attrs)
}

func (h *CustomAttributeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EntityType    string   `json:"entity_type"`
		AttributeKey  string   `json:"attribute_key"`
		DisplayName   string   `json:"display_name"`
		AttributeType string   `json:"attribute_type"`
		Options       []string `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.EntityType == "" || req.AttributeKey == "" || req.DisplayName == "" {
		Error(w, http.StatusBadRequest, "entity_type, attribute_key and display_name required")
		return
	}
	if req.AttributeType == "" {
		req.AttributeType = "text"
	}
	if req.Options == nil {
		req.Options = []string{}
	}
	attr, err := h.attrs.Create(r.Context(), req.EntityType, req.AttributeKey, req.DisplayName, req.AttributeType, req.Options)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, attr)
}

func (h *CustomAttributeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		DisplayName string   `json:"display_name"`
		Options     []string `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	attr, err := h.attrs.Update(r.Context(), id, req.DisplayName, req.Options)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, attr)
}

func (h *CustomAttributeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.attrs.Delete(r.Context(), id); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
