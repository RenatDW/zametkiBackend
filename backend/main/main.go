package main

import (
	"fmt"
	"net/http"
	"os"

	"notes-server/handlers"
	"notes-server/middleware"
	"notes-server/repository"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()

	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", ""),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", ""),
	)
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	addr := getEnv("ADDR", ":8080")
	storageDir := getEnv("STORAGE_DIR", "storage")

	// Создаём корневую директорию для файлов, если её нет
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Fatalw("failed to create storage dir", "err", err)
	}

	repo, err := repository.New(dsn)
	if err != nil {
		logger.Fatalw("failed to open database", "err", err)
	}

	h := &handlers.Handler{
		Logger:     logger,
		Repo:       repo,
		JWTSecret:  jwtSecret,
		StorageDir: storageDir,
	}

	r := mux.NewRouter()

	// Публичные маршруты
	r.HandleFunc("/api/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/api/login", h.Login).Methods(http.MethodPost)

	// Защищённые маршруты
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.Auth(jwtSecret))
	api.HandleFunc("/files", h.ListFiles).Methods(http.MethodGet)
	api.HandleFunc("/files", h.UploadFile).Methods(http.MethodPost)
	api.HandleFunc("/files/{uuid}", h.DownloadFile).Methods(http.MethodGet)
	api.HandleFunc("/files/{uuid}", h.DeleteFile).Methods(http.MethodDelete)

	logger.Infow("server starting", "addr", addr, "storage", storageDir)
	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Fatal(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
