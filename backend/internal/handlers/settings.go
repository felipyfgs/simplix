package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/simplix/api/internal/repository"
)

type SettingsHandler struct{ settings *repository.SettingsRepo }

func NewSettingsHandler(s *repository.SettingsRepo) *SettingsHandler { return &SettingsHandler{settings: s} }

func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settings.GetAll(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, settings)
}

func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	for k, v := range req {
		if err := h.settings.Set(r.Context(), k, v); err != nil {
			Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	JSON(w, http.StatusOK, map[string]string{"message": "updated"})
}
