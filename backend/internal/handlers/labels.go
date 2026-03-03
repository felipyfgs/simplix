package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/repository"
)

type LabelHandler struct{ labels *repository.LabelRepo }

func NewLabelHandler(labels *repository.LabelRepo) *LabelHandler { return &LabelHandler{labels: labels} }

func (h *LabelHandler) List(w http.ResponseWriter, r *http.Request) {
	labels, err := h.labels.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, labels)
}

func (h *LabelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Color       string  `json:"color"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		Error(w, http.StatusBadRequest, "name required")
		return
	}
	if req.Color == "" {
		req.Color = "#6B7280"
	}
	label, err := h.labels.Create(r.Context(), req.Name, req.Color, req.Description)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, label)
}

func (h *LabelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Name        string  `json:"name"`
		Color       string  `json:"color"`
		Description *string `json:"description"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	label, err := h.labels.Update(r.Context(), id, req.Name, req.Color, req.Description)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, label)
}

func (h *LabelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	h.labels.Delete(r.Context(), id)
	w.WriteHeader(http.StatusNoContent)
}
