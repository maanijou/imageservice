package image

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Image struct {
	gorm.Model
	ID       uuid.UUID `json:"uuid"`
	FileName string    `json:"file_name"`
	FileSize int       `json:"file_size"`
	MimeType string    `json:"mime_type"`
}
