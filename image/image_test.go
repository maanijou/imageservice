package image_test

import (
	"log"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/maanijou/imageservice/database"
	"github.com/maanijou/imageservice/image"
)

// setup function to create virtual server.

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		teardown()
		os.Exit(-1)
		return
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() error {
	log.Println("Setup image test")
	err := database.InitDatabase("test")
	if err != nil {
		return err
	}
	err = database.GlobalDB.AutoMigrate(&image.Image{})
	if err != nil {
		return err
	}
	database.GlobalDB.AutoMigrate(&image.Image{})
	return nil
}

func TestDbImage(t *testing.T) {
	id, err := uuid.NewUUID()
	if err != nil {
		t.Error("Error creating new UUID for the image")
	}
	img := image.Image{
		ID:       id,
		FileName: "test.png",
		FileSize: 100,
		MimeType: "image/png",
	}
	err = img.CreateImage()
	if err != nil {
		t.Errorf("Error adding image to the database\n")
	}
	got, err := image.GetImageByID(id)
	if err != nil {
		t.Errorf("Error getting the image\n")
	}
	if got.FileName != img.FileName {
		t.Errorf("Error getting the image\n")
	}
}

func TestImageUpload(t *testing.T) {
	t.Error("Not implemented")
}

func TestMultipleImageUpload(t *testing.T) {
	t.Error("Not implemented")
}

func TestImageDownload(t *testing.T) {
	t.Error("Not implemented")
}

func TestImageFormatconversion(t *testing.T) {
	t.Error("Not implemented")
}

func teardown() {
	err := os.Remove("test.db")
	if err != nil {
		log.Println("Error removing test.db file")
	}
	log.Println("cleanup image test")
}
