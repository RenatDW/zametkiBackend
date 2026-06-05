package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"notes-server/middleware"
	"notes-server/models"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

const maxUploadSize = 50 << 20 // 50 МБ

// ListFiles — GET /api/files
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	files, err := h.Repo.ListFiles(userID)
	if err != nil {
		h.Logger.Errorw("list files error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]models.FileResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, toFileResponse(f))
	}
	h.writeJSON(w, http.StatusOK, resp)
}

// UploadFile — POST /api/files
// Content-Type: multipart/form-data, поле "file"
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.writeError(w, http.StatusBadRequest, "file too large or invalid form data")
		return
	}

	src, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "field 'file' is required")
		return
	}
	defer src.Close()

	fileUUID := uuid.NewString()

	// Директория пользователя: <storage_dir>/users/<userID>/
	userDir := filepath.Join(h.StorageDir, "users", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		h.Logger.Errorw("mkdir error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Файл на диске называется просто UUID — без оригинального имени (защита от path traversal)
	storagePath := filepath.Join(userDir, fileUUID)
	dst, err := os.Create(storagePath)
	if err != nil {
		h.Logger.Errorw("create file error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, src)
	if err != nil {
		os.Remove(storagePath)
		h.Logger.Errorw("write file error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	meta := &models.File{
		UUID:         fileUUID,
		UserID:       userID,
		OriginalName: header.Filename,
		Size:         written,
		StoragePath:  storagePath,
	}
	if err := h.Repo.SaveFileMeta(meta); err != nil {
		os.Remove(storagePath)
		h.Logger.Errorw("save meta error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.writeJSON(w, http.StatusCreated, toFileResponse(*meta))
}

// DownloadFile — GET /api/files/{uuid}
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	fileUUID := mux.Vars(r)["uuid"]

	meta, err := h.Repo.GetFile(fileUUID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.writeError(w, http.StatusNotFound, "file not found")
		return
	} else if err != nil {
		h.Logger.Errorw("get file meta error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	f, err := os.Open(meta.StoragePath)
	if err != nil {
		h.Logger.Errorw("open file error", "err", err, "path", meta.StoragePath)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", `attachment; filename="`+filepath.Base(meta.OriginalName)+`"`)
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, meta.OriginalName, meta.CreatedAt, f)
}

// DeleteFile — DELETE /api/files/{uuid}
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	fileUUID := mux.Vars(r)["uuid"]

	meta, err := h.Repo.DeleteFile(fileUUID, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.writeError(w, http.StatusNotFound, "file not found")
		return
	} else if err != nil {
		h.Logger.Errorw("delete file db error", "err", err)
		h.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := os.Remove(meta.StoragePath); err != nil && !os.IsNotExist(err) {
		h.Logger.Warnw("failed to remove file from disk", "err", err, "path", meta.StoragePath)
	}

	w.WriteHeader(http.StatusNoContent)
}

func toFileResponse(f models.File) models.FileResponse {
	return models.FileResponse{
		UUID:      f.UUID,
		Name:      f.OriginalName,
		Size:      f.Size,
		CreatedAt: f.CreatedAt.Format(time.RFC3339),
	}
}
