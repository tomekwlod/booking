package main

// https://stackoverflow.com/questions/34177137/stream-file-upload-to-aws-s3-using-go
// https://medium.com/spankie/upload-images-to-aws-s3-bucket-in-a-golang-web-application-2612bea70dd8
// https://docs.digitalocean.com/products/spaces/resources/s3-sdk-examples/
// https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
// https://codepen.io/PerfectIsShit/pen/zogMXP -- progress bar done on a frontend

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/tomekwlod/utils/env"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func dbtestHandler(w http.ResponseWriter, r *http.Request) {

	err := pconn.Ping()

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

	// parse file field
	mp, err := reader.NextPart()
	if err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if mp.FormName() != "profile_picture" {
		http.Error(w, "profile_picture is expected", http.StatusBadRequest)
		return
	}

	// buf := bufio.NewReader(mp)
	// sniff, _ := buf.Peek(512)
	// contentType := http.DetectContentType(sniff)
	// if contentType != "application/zip" {
	// 	http.Error(w, "file type not allowed", http.StatusBadRequest)
	// 	return
	// }

	awsConfig := &aws.Config{
		Endpoint: aws.String(env.Env("DO_S3_ENDPOINT", "https://nyc3.digitaloceanspaces.com")),
		Region:   aws.String(env.Env("DO_S3_REGION", "us-east-1")),

		Credentials: credentials.NewStaticCredentials(
			env.Env("DO_S3_KEY", ""),
			env.Env("DO_S3_SECRET", ""),
			"",
		),
	}

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(awsConfig))

	fileName, err := uploadFileToS3(sess, mp)

	if err != nil {
		fmt.Fprintf(w, "Could not upload file %v", err)
		return
	}

	fmt.Fprintf(w, "Image uploaded successfully to: %v", fileName)
}

// uploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func uploadFileToS3(sess *session.Session, file *multipart.Part) (string, error) {
	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 2            // default is 5
	})

	fileext := filepath.Ext(file.FileName())
	filename := file.FileName()[0 : len(file.FileName())-len(fileext)]

	// create a unique file name for the file
	tempFileName := path.Join(
		env.Env("DO_S3_DIR", ""),
		fmt.Sprintf("%s_%s%s", filename, uuid.New().String(), filepath.Ext(file.FileName())),
	)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(env.Env("DO_S3_BUCKET", "booking")),
		Key:    aws.String(tempFileName),
		Body:   file,
	})

	// in case it fails to upload
	if err != nil {
		return "", err
	}

	return result.Location, nil
}
