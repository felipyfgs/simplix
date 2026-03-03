package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/simplix/api/internal/domain"
	"github.com/simplix/api/internal/middleware"
	"github.com/simplix/api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users     *repository.UserRepo
	jwtSecret string
}

func NewAuthHandler(users *repository.UserRepo, secret string) *AuthHandler {
	return &AuthHandler{users: users, jwtSecret: secret}
}

func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}

	user, err := h.users.FindByEmail(r.Context(), req.Email)
	if err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		Error(w, http.StatusInternalServerError, "token error")
		return
	}

	user.PasswordHash = ""
	JSON(w, http.StatusOK, map[string]any{
		"token": signed,
		"user":  user,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		Error(w, http.StatusNotFound, "user not found")
		return
	}
	user.PasswordHash = ""
	JSON(w, http.StatusOK, user)
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.Email == "" || req.Password == "" {
		Error(w, http.StatusBadRequest, "name, email and password are required")
		return
	}
	if len(req.Password) < 8 {
		Error(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Error(w, http.StatusInternalServerError, "hash error")
		return
	}
	user, err := h.users.Create(r.Context(), req.Name, req.Email, string(hash), domain.RoleAgent)
	if err != nil {
		Error(w, http.StatusConflict, "email already in use")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		Error(w, http.StatusInternalServerError, "token error")
		return
	}

	user.PasswordHash = ""
	JSON(w, http.StatusCreated, map[string]any{
		"token": signed,
		"user":  user,
	})
}

func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"message": "signed out"})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		Error(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Name         string  `json:"name"`
		Availability string  `json:"availability"`
		AvatarURL    *string `json:"avatar_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}
	avail := domain.Availability(req.Availability)
	if avail == "" {
		avail = domain.AvailabilityOffline
	}
	user, err := h.users.UpdateProfile(r.Context(), userID, req.Name, avail, req.AvatarURL)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.PasswordHash = ""
	JSON(w, http.StatusOK, user)
}
