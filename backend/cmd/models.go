package models

import "gorm.io/gorm"

// User — пользователь системы
type User struct {
	gorm.Model
	Username     string  `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string  `gorm:"not null"             json:"-"`
	Files        []File  `gorm:"foreignKey:UserID"    json:"-"`
}

// File — метаданные файла; сам файл лежит на диске по пути StoragePath
type File struct {
	gorm.Model
	UUID        string `gorm:"uniqueIndex;not null" json:"uuid"`
	UserID      uint   `gorm:"not null;index"       json:"user_id"`
	OriginalName string `gorm:"not null"            json:"name"`
	Size        int64  `gorm:"not null"             json:"size"`
	StoragePath string `gorm:"not null"             json:"-"` // путь на диске, клиенту не отдаём
}

// --- DTOs ---

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// FileResponse — то, что получает клиент в списке или после загрузки
type FileResponse struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}
