package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"notes-server/middleware"
	"notes-server/models"

	"github.com/gorilla/mux"
)

func (h *Handler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	notes, err := h.Repo.GetNotesByUserID(userID)
	if err != nil {
		h.Logger.Errorw("get notes error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if notes == nil {
		notes = []models.Note{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (h *Handler) CreateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.CreateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := h.Repo.CreateNote(userID, req.Title, req.Content)
	if err != nil {
		h.Logger.Errorw("create note error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	noteID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	var req models.UpdateNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	note, err := h.Repo.UpdateNote(noteID, userID, req.Title, req.Content)
	if err == sql.ErrNoRows {
		h.writeError(w, http.StatusNotFound, "note not found")
		return
	} else if err != nil {
		h.Logger.Errorw("update note error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

func (h *Handler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	noteID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid note id")
		return
	}

	if err := h.Repo.DeleteNote(noteID, userID); err == sql.ErrNoRows {
		h.writeError(w, http.StatusNotFound, "note not found")
		return
	} else if err != nil {
		h.Logger.Errorw("delete note error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
