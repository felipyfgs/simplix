package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/middleware"
	"github.com/simplix/api/internal/repository"
	"github.com/simplix/api/internal/service"
)

type NoteHandler struct {
	notes  *repository.NoteRepo
	broker *service.SSEBroker
}

func NewNoteHandler(notes *repository.NoteRepo, broker *service.SSEBroker) *NoteHandler {
	return &NoteHandler{notes: notes, broker: broker}
}

func (h *NoteHandler) List(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}
	notes, err := h.notes.List(r.Context(), contactID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if notes == nil {
		notes = []domain.Note{}
	}
	JSON(w, http.StatusOK, notes)
}

func (h *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		Error(w, http.StatusBadRequest, "content required")
		return
	}
	userID, _ := middleware.GetUserID(r)
	note, err := h.notes.Create(r.Context(), contactID, &userID, req.Content)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.broker.BroadcastAll(domain.SSEEvent{Type: "note.created", Payload: note})
	JSON(w, http.StatusCreated, note)
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid contact id")
		return
	}
	noteID, err := uuid.Parse(chi.URLParam(r, "nid"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid note id")
		return
	}
	if err := h.notes.Delete(r.Context(), noteID, contactID); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}
