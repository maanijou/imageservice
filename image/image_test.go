package image_test

import (
	"log"
	"os"
	"testing"

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
	err = os.ErrClosed
	if err != nil {
		return err
	}
	database.GlobalDB.AutoMigrate(&image.Image{})
	return nil
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
