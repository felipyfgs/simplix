package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
)

type ContactHandler struct {
	contacts *repository.ContactRepo
}

func NewContactHandler(contacts *repository.ContactRepo) *ContactHandler {
	return &ContactHandler{contacts: contacts}
}

func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := ParsePage(r)
	q := r.URL.Query()

	// Parse filter=field:op:value params (multiple allowed)
	var conditions []repository.FilterCondition
	for _, raw := range q["filter"] {
		parts := splitN(raw, ":", 3)
		if len(parts) == 3 {
			conditions = append(conditions, repository.FilterCondition{
				Field: parts[0], Op: parts[1], Value: parts[2],
			})
		} else if len(parts) == 2 { // present / not_present have no value
			conditions = append(conditions, repository.FilterCondition{
				Field: parts[0], Op: parts[1],
			})
		}
	}

	filter := repository.ContactFilter{
		Query:      q.Get("q"),
		Status:     q.Get("status"),
		Label:      q.Get("label"),
		Conditions: conditions,
		Page:       page,
		Limit:      limit,
	}
	contacts, total, err := h.contacts.List(r.Context(), filter)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if contacts == nil {
		contacts = []domain.Contact{}
	}
	JSON(w, http.StatusOK, domain.PagedResult[domain.Contact]{
		Data:       contacts,
		Pagination: domain.Pagination{Page: page, Limit: limit, Total: total},
	})
}

func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name             string         `json:"name"`
		Email            *string        `json:"email"`
		Phone            *string        `json:"phone"`
		Company          *string        `json:"company"`
		CustomAttributes map[string]any `json:"custom_attributes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}
	contact, err := h.contacts.Create(r.Context(), &req.Name, req.Email, req.Phone, req.Company, req.CustomAttributes)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, contact)
}

func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	contact, err := h.contacts.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "contact not found")
		return
	}
	labels, _ := h.contacts.GetLabels(r.Context(), id)
	contact.Labels = labels
	JSON(w, http.StatusOK, contact)
}

func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	// Handle custom_attributes separately
	if ca, ok := body["custom_attributes"]; ok {
		if caMap, ok := ca.(map[string]any); ok {
			if err := h.contacts.UpdateCustomAttributes(r.Context(), id, caMap); err != nil {
				Error(w, http.StatusInternalServerError, err.Error())
				return
			}
		}
		delete(body, "custom_attributes")
	}

	if len(body) > 0 {
		if _, err := h.contacts.Update(r.Context(), id, body); err != nil {
			Error(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	contact, _ := h.contacts.FindByID(r.Context(), id)
	JSON(w, http.StatusOK, contact)
}

func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.contacts.Delete(r.Context(), id); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ContactHandler) GetLabels(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	labels, err := h.contacts.GetLabels(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if labels == nil {
		labels = []domain.Label{}
	}
	JSON(w, http.StatusOK, labels)
}

func (h *ContactHandler) SetLabels(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		LabelIDs []uuid.UUID `json:"label_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.contacts.SetLabels(r.Context(), id, req.LabelIDs); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	labels, _ := h.contacts.GetLabels(r.Context(), id)
	if labels == nil {
		labels = []domain.Label{}
	}
	JSON(w, http.StatusOK, labels)
}
