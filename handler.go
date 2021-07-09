package main

// https://stackoverflow.com/questions/34177137/stream-file-upload-to-aws-s3-using-go
// https://medium.com/spankie/upload-images-to-aws-s3-bucket-in-a-golang-web-application-2612bea70dd8
// https://docs.digitalocean.com/products/spaces/resources/s3-sdk-examples/
// https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
// https://codepen.io/PerfectIsShit/pen/zogMXP -- progress bar done on a frontend
// https://github.com/aws/aws-sdk-go/commit/50ba1dfe47983b15b160b66c730a3b93d2961f8e -- progress to stdout

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/tomekwlod/booking/internal/s3"
	"github.com/tomekwlod/utils/env"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func dbtestHandler(w http.ResponseWriter, r *http.Request) {

	err := dbConn.Ping()

	if err != nil {

		log.Printf("Error while pinging DB, %v", err)
		writeError(w, errInternalServer)
	}

	w.Write([]byte("ok"))
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {

	type jsonRes struct {
		Username string `json:"username"`
		Code     string `json:"code"`
		ID       int    `json:"id"`
	}

	j := &jsonRes{"Tomek27", "39s94jr99c", 1}

	writeJSON(w, j, http.StatusOK)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// function body of a http.HandlerFunc
	r.Body = http.MaxBytesReader(w, r.Body, 25<<20+1024) // 25Mb
	reader, err := r.MultipartReader()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doClient := s3.New(
		env.Env("DO_S3_ENDPOINT", "https://nyc3.digitaloceanspaces.com"),
		env.Env("DO_S3_REGION", "us-east-1"),
		env.Env("DO_S3_KEY", ""),
		env.Env("DO_S3_SECRET", ""),
		env.Env("DO_S3_BUCKET", "booking"),
		env.Env("DO_S3_DIR", ""),
	)

	destinationFile, err := uploadFile(reader, "profile_picture", doClient)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Image uploaded successfully to: %v", destinationFile)
}

func uploadFile(reader *multipart.Reader, formName string, u Uploader) (string, error) {

	// parse file field
	mp, err := reader.NextPart()

	if err != nil && err != io.EOF {

		return "", err
	}

	if mp.FormName() != formName {

		return "", err
	}

	return u.Upload(mp)
}

type Uploader interface {
	Upload(mp *multipart.Part) (destination string, err error)
}
