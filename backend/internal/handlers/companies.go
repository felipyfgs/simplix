package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
)

type CompanyHandler struct{ companies *repository.CompanyRepo }

func NewCompanyHandler(companies *repository.CompanyRepo) *CompanyHandler {
	return &CompanyHandler{companies: companies}
}

func (h *CompanyHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := ParsePage(r)
	q := r.URL.Query().Get("q")
	sort := r.URL.Query().Get("sort")
	companies, total, err := h.companies.List(r.Context(), q, page, limit, sort)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if companies == nil {
		companies = []domain.Company{}
	}
	JSON(w, http.StatusOK, domain.PagedResult[domain.Company]{
		Data:       companies,
		Pagination: domain.Pagination{Page: page, Limit: limit, Total: total},
	})
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string  `json:"name"`
		Domain      *string `json:"domain"`
		Phone       *string `json:"phone"`
		Website     *string `json:"website"`
		Industry    *string `json:"industry"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}
	company, err := h.companies.Create(r.Context(), req.Name, req.Domain, req.Phone, req.Website, req.Industry, req.Description)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, company)
}

func (h *CompanyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	company, err := h.companies.FindByID(r.Context(), id)
	if err != nil {
		Error(w, http.StatusNotFound, "company not found")
		return
	}
	JSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	company, err := h.companies.Update(r.Context(), id, body)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.companies.Delete(r.Context(), id); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CompanyHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	contacts, err := h.companies.ListContacts(r.Context(), id)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if contacts == nil {
		contacts = []domain.Contact{}
	}
	JSON(w, http.StatusOK, contacts)
}
