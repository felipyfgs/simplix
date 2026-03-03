package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct{ users *repository.UserRepo }

func NewUserHandler(users *repository.UserRepo) *UserHandler { return &UserHandler{users: users} }

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	JSON(w, http.StatusOK, users)
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string      `json:"name"`
		Email    string      `json:"email"`
		Password string      `json:"password"`
		Role     domain.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "hash error")
		return
	}
	if req.Role == "" {
		req.Role = domain.RoleAgent
	}
	user, err := h.users.Create(r.Context(), req.Name, req.Email, string(hash), req.Role)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.PasswordHash = ""
	JSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Name         string             `json:"name"`
		Availability domain.Availability `json:"availability"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Availability == "" {
		req.Availability = domain.AvailabilityOnline
	}
	user, err := h.users.Update(r.Context(), id, req.Name, req.Availability)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.PasswordHash = ""
	JSON(w, http.StatusOK, user)
}
