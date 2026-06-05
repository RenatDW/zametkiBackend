package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"notes-server/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		h.writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if len(req.Password) < 6 {
		h.writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.Errorw("bcrypt error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := h.Repo.CreateUser(req.Username, string(hash)); err != nil {
		h.writeError(w, http.StatusConflict, "username already taken")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.Repo.GetUserByUsername(req.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	} else if err != nil {
		h.Logger.Errorw("db error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		h.writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		h.Logger.Errorw("jwt sign error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.writeJSON(w, http.StatusOK, models.LoginResponse{Token: tokenStr})
}
