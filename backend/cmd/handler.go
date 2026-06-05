package handlers

import (
	"encoding/json"
	"net/http"

	"notes-server/repository"

	"go.uber.org/zap"
)

type Handler struct {
	Logger    *zap.SugaredLogger
	Repo      *repository.Repository
	JWTSecret string
}

func (h *Handler) writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
