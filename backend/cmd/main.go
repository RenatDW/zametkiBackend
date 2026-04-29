package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.SugaredLogger
	repo   int // todo Реализовать взаимодействие с бд
}

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func(zapLogger *zap.Logger) {
		err = zapLogger.Sync()
		if err != nil {
			panic(err)
		}
	}(zapLogger)
	logger := zapLogger.Sugar()

	h := &Handler{
		Logger: logger,
		repo:   0,
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/login", h.SyncNote)
	if err := http.ListenAndServe(":8080", r); err != nil {
		logger.Fatal(err)
	}

}

func (h *Handler) SyncNote(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("/api/login")
	// var req SyncRequest
	// json.NewDecoder(r.Body).Decode(&req)

	// updateBytes, _ := base64.StdEncoding.DecodeString(req.Snapshot)

	// err := h.repo.SaveUpdate(req.NoteID, updateBytes, req.ClientID)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	// lastID := h.repo.GetLastUpdateID(req.NoteID)

	// json.NewEncoder(w).Encode(SyncResponse{
	// 	Status:       "ok",
	// 	LastUpdateID: lastID,
	// })
}
