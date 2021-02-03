package image_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdImage "image"
	"image/color"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/maanijou/imageservice/database"
	"github.com/maanijou/imageservice/image"
)

// setup function to create virtual server.

var sm *mux.Router
var clearThese []string

func TestMain(m *testing.M) {
	route, err := setup()
	sm = route
	if err != nil {
		teardown()
		os.Exit(-1)
		return
	}
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() (*mux.Router, error) {
	log.Println("Setup image test")
	os.Chdir("..")
	sm := mux.NewRouter().StrictSlash(true) // ignoring trailing slash
	sm = sm.PathPrefix("/api/v1/").Subrouter()
	image.SetupRoutes(sm)
	err := database.InitDatabase("test")
	if err != nil {
		return sm, err
	}
	err = database.GlobalDB.AutoMigrate(&image.Image{})
	if err != nil {
		return sm, err
	}
	_, err = sampleImage()
	if err != nil {
		return sm, err
	}
	return sm, nil
}

func teardown() {
	err := os.Remove("test.db")
	if err != nil {
		log.Println("Error removing test.db file")
	}

	for _, im := range clearThese {
		err = os.Remove("uploads/" + im)
		if err != nil {
			log.Println("Error removing sample file")
		}
	}
	log.Println("cleanup image test")
}

func sampleImage(inputID ...string) (stdImage.Image, error) {
	myID := "269e6804-7b8e-4504-bf7e-fbf2d2bbc926"
	if len(inputID) > 0 {
		myID = inputID[0]
	}
	id, err := uuid.ParseBytes([]byte(myID))
	if err != nil {
		return nil, err
	}

	img := stdImage.NewRGBA(stdImage.Rect(0, 0, 100, 50))

	// Draw a red dot at (2, 3)
	img.Set(2, 3, color.RGBA{255, 0, 0, 255})

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, err
	}

	im := image.Image{
		ID:       id,
		FileName: "test.png",
		FileSize: buf.Len(),
		MimeType: "image/png",
	}
	err = im.CreateImage()
	if err != nil {
		return nil, err
	}
	// Save to out.png
	f, err := os.OpenFile(
		filepath.Join(
			"uploads",
			fmt.Sprintf("%s.png", im.ID.String()),
			// "test",
		),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	defer f.Close()
	err = png.Encode(f, img)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	clearThese = append(clearThese, id.String()+".png")
	return img, nil
}

func TestDbImage(t *testing.T) {
	id, err := uuid.NewRandom()
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

func TestImageDownload(t *testing.T) {
	scenarios := []struct {
		path   string
		expect string
		status int
	}{
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926", "image/png", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc925", "application/json", 404},
		{"/api/v1/image/269e680", "text/plain", 404},
		{"/api/v1/image/", "", 405},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=png", "image/png", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=gif", "image/gif", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=jpg", "image/jpeg", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=jpeg", "image/jpeg", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=", "image/png", 200},
		{"/api/v1/image/269e6804-7b8e-4504-bf7e-fbf2d2bbc926?ext=test", "application/json", 400},
	}
	for _, s := range scenarios {
		req, err := http.NewRequest(
			"GET",
			s.path,
			nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.Handler(sm)
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != s.status {
			t.Errorf("handler returned wrong status code: got %v want %v %s",
				rr.Code, s.status, s.path)
		}
		if !strings.Contains(rr.Header().Get("Content-Type"), s.expect) {
			t.Errorf("handler returned wrong content type: got %v want %v",
				rr.Header().Get("Content-Type"), s.expect)
		}
	}
}

func TestImageUpload(t *testing.T) {

	img, err := sampleImage("583e6804-764e-4504-367e-fb66d2bbc934")
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	imageBuf := new(bytes.Buffer)
	err = png.Encode(imageBuf, img)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("myFiles", "test")
	if err != nil {
		writer.Close()
		t.Error(err)
	}
	io.Copy(part, imageBuf)
	writer.Close()
	req := httptest.NewRequest(
		"POST",
		"/api/v1/image",
		body)

	// req.Header.Set("Content-Type", "multipart/form-data")//?
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler := http.Handler(sm)
	handler.ServeHTTP(rr, req)
	var gotImage []image.Image
	err = json.Unmarshal(rr.Body.Bytes(), &gotImage)
	if err != nil {
		t.Error("Error unmarshaling results")
		t.FailNow()
	}
	clearThese = append(clearThese, gotImage[0].ID.String()+".png")
	if status := rr.Code; status != 202 {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, 202)
		t.FailNow()
	}
	if gotImage[0].FileSize != 204 {
		t.Errorf("handler returned wrong file size: got %v want %v",
			gotImage[0].FileSize, 204)
		t.FailNow()
	}
}

// TODO More tests:
// Check for file size
// Check multiple file uploads
// Check for wrong file contents
// Benchmarking conversion
// Test coverage
