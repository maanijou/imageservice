package image

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Image struct {
	ID       uuid.UUID `gorm:"primaryKey" json:"id"`
	FileName string    `json:"file_name"`
	FileSize int       `json:"file_size"`
	MimeType string    `json:"mime_type"`
}

func (u *Image) AfterDelete(tx *gorm.DB) error {
	if tx.RowsAffected > 0 {
		err := os.Remove(
			filepath.Join(
				"uploads",
				u.ID.String(),
			))
		if err != nil {
			log.Println(errors.Wrap(err, "Error removing file in AfterDelete Hook"))
			return err
		}
	}
	return nil
}

func ConvertImage(data []byte, fromMimeType string, toMimeType string) ([]byte, error) {
	var img image.Image
	var err error
	switch fromMimeType {
	case "image/png":
		img, err = png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode png")
		}
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode jpeg")
		}
	case "image/jpg":
		img, err = jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode jpeg")
		}
	case "image/gif":
		img, err = gif.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode gif")
		}
	default:
		img, err = png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, errors.Wrap(err, "unable to decode png")
		}
	}

	buf := new(bytes.Buffer)
	switch toMimeType {
	case "image/png":
		err := png.Encode(buf, img)
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode png")
		}
	case "image/jpeg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode jpg")
		}
	case "image/jpg":
		err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode jpg")
		}
	case "image/gif":
		err := gif.Encode(buf, img, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode gif")
		}
	default:
		err := png.Encode(buf, img)
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode png")
		}
	}
	return buf.Bytes(), nil
}
