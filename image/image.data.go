package image

import (
	"log"

	"github.com/maanijou/imageservice/database"
	"github.com/pkg/errors"

	"github.com/google/uuid"
)

func (image *Image) CreateImage() error {
	result := database.GlobalDB.Create(&image)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetImageByID(ID uuid.UUID) (*Image, error) {
	var image Image
	result := database.GlobalDB.Where("id = ?", ID.String()).First(&image)

	if result.Error != nil {
		log.Println(
			errors.Wrap(result.Error, "Error getting image from db"),
		)
		return nil, result.Error
	}
	return &image, nil
}
