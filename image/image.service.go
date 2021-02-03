package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"log"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func getFile(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	vars := mux.Vars(r)
	ext := "image/png"
	switch queries.Get("ext") {
	case "":
	case "png":
		ext = "image/png"
	case "jpg":
		ext = "image/jpeg"
	case "jpeg":
		ext = "image/jpeg"
	case "gif":
		ext = "image/gif"
	default:
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"error": "This image extention is not supported. Supported extends are: png, jpg, jpeg, and gif"}`,
		))
		return
	}
	id, err := uuid.Parse(vars["id"])
	// Gorilla mux should not let this happen, but anyway...
	if err != nil {
		log.Println(errors.Wrap(err, "Error in getting the image id"))
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(
			`{"error": "You have to provide the id field"}`,
		))
		return
	}

	image, err := GetImageByID(id)
	if err != nil {
		log.Println(errors.Wrap(err, "Requested id does not exist"))
		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(
			`{"error": "No image found with given id"}`,
		))
		return
	}
	contentType := "png"
	switch image.MimeType {
	case "image/jpeg":
		contentType = "jpg"
	case "image/jpg":
		contentType = "jpg"
	case "image/gif":
		contentType = "gif"
	default:
		contentType = "png"
	}
	data, err := ioutil.ReadFile(
		filepath.Join(
			"uploads",
			fmt.Sprintf("%s.%s", id, contentType),
		))
	if err != nil {
		log.Println(errors.Wrap(err, "Error opening the image file"))

		w.Header().Add("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(
			`{"error": "No image found with given id"}`,
		))
		return
	}

	if image.MimeType != ext {
		data, err = ConvertImage(data, image.MimeType, ext)
		if err != nil {
			log.Println("Error converting image: ", err)
			w.Header().Add("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(
				`{"error": "Error converting image"}`,
			))
			return
		}
	}
	w.Header().Add("Content-Type", fmt.Sprintf("%s; charset=UTF-8", ext))
	w.Write(data)

}
func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	fhs := r.MultipartForm.File["myFiles"]

	var respond []Image
	var wg sync.WaitGroup
	for _, fh := range fhs {
		wg.Add(1)
		go func(fh *multipart.FileHeader, wg *sync.WaitGroup) {
			defer wg.Done()
			if fh.Size > 10<<20 {
				log.Println(errors.New("File size is bigger than 10MB limit"))
				return
			}
			f, err := fh.Open()
			if err != nil {
				log.Println(errors.Wrap(err, "Error Retrieving the File"))
				return
			}
			defer f.Close()

			// read all of the contents of our uploaded file into a
			// byte array
			fileBytes, err := ioutil.ReadAll(f)
			if err != nil {
				log.Println(errors.Wrap(err, "Error reading image file"))
				return
			}
			id, err := uuid.NewRandom()
			if err != nil {
				log.Println(errors.Wrap(err, "Error creating UUID"))
				return
			}
			contentType := http.DetectContentType(fileBytes)
			ext := "png"
			switch contentType {
			case "image/jpeg":
				ext = "jpg"
			case "image/jpg":
				ext = "jpg"
			case "image/gif":
				ext = "gif"
			default:
				ext = "png"
			}
			var image *Image = &Image{
				ID:       id,
				FileName: fh.Filename,
				FileSize: int(fh.Size),
				MimeType: contentType,
			}
			err = image.CreateImage()
			if err != nil {
				log.Println(errors.Wrap(err, "Error creating image"))
				return
			}
			file, err := os.OpenFile(
				filepath.Join(
					"uploads",
					fmt.Sprintf("%s.%s", id, ext),
				), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				log.Println(errors.Wrap(err, "Error creating the File"))
				return
			}
			defer file.Close()
			respond = append(respond, *image)
			// write this byte array to our temporary file
			file.Write(fileBytes)
		}(fh, &wg)
		wg.Wait()
	}

	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusAccepted)
	if respond == nil {
		w.Write([]byte(
			`[]`,
		))
	} else {
		json.NewEncoder(w).Encode(respond)
	}
}

// SetupRoutes is used to setup necessary routes for image service
func SetupRoutes(sm *mux.Router) {
	postRouter := sm.Methods(http.MethodPost).Subrouter()
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	postRouter.HandleFunc("/image", uploadFile)
	getRouter.HandleFunc("/image/{id:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}}", getFile)
}
