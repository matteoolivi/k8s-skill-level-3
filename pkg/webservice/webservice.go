package webservice

import (
	"fmt"
	"log"
	"net/http"

	"github.com/matteoolivi/img-blurring-exercise/pkg/helper/pg"
	"github.com/matteoolivi/img-blurring-exercise/pkg/helper/rabbitmq"
	"github.com/matteoolivi/img-blurring-exercise/pkg/helper/s3"
)

// TODO: Make logging non-embarassing.

// TODO: Revamp error messages and HTTP status codes.

// TODO: Add repair logic if this process crashes in the middle of an upload (e.g if there's an
// upload to S3 but there's a crash before the metadata is uploaded to PostgreSQL, on restart the
// metadata should be uploaded to PostgreSQL). Such logic might, but does not have to, reside in
// this process.

// TODO: Ensure S3 bucket existence.

// TODO: Ensure Postgres DB and table existence.

const (
	host = ""
	port = 8080
)

func ListenAndServe() {
	http.HandleFunc("/picture", func(w http.ResponseWriter, r *http.Request) {
		handlePicture(w, r)
	})

	log.Printf("Listening on port %d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil))
}

func handlePicture(w http.ResponseWriter, r *http.Request) {
	// TODO: Factor out this check to a wrapper function.
	if r.Method != http.MethodPost {
		handleInvalidMethod(r.Method, w)
		return
	}

	imageName, imageURL, err := uploadToS3(r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Succesfully uploaded image %s on S3. URL: %s", imageName, imageURL)

	err = pg.StoreMetadata(imageName, imageURL)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("Succesfully stored pair (name = %s, url = %s) for image %s in metadata DB.",
		imageName,
		imageURL,
		imageName)

	err = rabbitmq.EnqueueURL(imageName, imageURL)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Image " +
		imageName +
		" has been uploaded to AWS S3 at URL " +
		imageURL +
		" and is waiting to be blurred."))
}

func handleInvalidMethod(method string, w http.ResponseWriter) {
	log.Printf("Received request with unspported method %s", method)

	msg := []byte("Request method is " + method + ", server only supports " + http.MethodPost)

	w.WriteHeader(http.StatusMethodNotAllowed)
	if _, err := w.Write(msg); err != nil {
		log.Printf("Failed to write response \"%s\" to client: %s", msg, err)
	}
}

func uploadToS3(r *http.Request) (string, string, error) {
	image, metadata, err := r.FormFile("image")
	if err != nil {
		return "", "", err
	}
	defer image.Close()

	imageName := metadata.Filename
	imageURL, err := s3.UploadImage(imageName, image)
	if err != nil {
		return "", "", err
	}

	return imageName, imageURL, nil
}
