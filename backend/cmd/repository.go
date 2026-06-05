package repository

import (
	"errors"
	"notes-server/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.File{}); err != nil {
		return nil, err
	}
	return &Repository{db: db}, nil
}

// --- Users ---

func (r *Repository) CreateUser(username, passwordHash string) error {
	return r.db.Create(&models.User{
		Username:     username,
		PasswordHash: passwordHash,
	}).Error
}

func (r *Repository) GetUserByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.db.Where("username = ?", username).First(&u).Error
	return &u, err
}

// --- Files ---

// SaveFileMeta сохраняет запись о файле в БД
func (r *Repository) SaveFileMeta(f *models.File) error {
	return r.db.Create(f).Error
}

// ListFiles возвращает метаданные всех файлов пользователя
func (r *Repository) ListFiles(userID uint) ([]models.File, error) {
	var files []models.File
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&files).Error
	return files, err
}

// GetFile возвращает метаданные файла, проверяя владельца
func (r *Repository) GetFile(uuid string, userID uint) (*models.File, error) {
	var f models.File
	err := r.db.Where("uuid = ? AND user_id = ?", uuid, userID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &f, err
}

// DeleteFile удаляет запись из БД; удаление файла с диска — на стороне хендлера
func (r *Repository) DeleteFile(uuid string, userID uint) (*models.File, error) {
	f, err := r.GetFile(uuid, userID)
	if err != nil {
		return nil, err
	}
	return f, r.db.Delete(f).Error
}
