package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

// UploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	fileext := filepath.Ext(fileHeader.Filename)
	filename := fileHeader.Filename[0 : len(fileHeader.Filename)-len(fileext)]

	// create a unique file name for the file
	tempFileName := path.Join(
		env.Env("DO_S3_DIR", ""),
		fmt.Sprintf("%s_%s%s", filename, uuid.New().String(), filepath.Ext(fileHeader.Filename)),
	)

	// config settings: this is where you choose the bucket,
	// filename, content-type and storage class of the file
	// you're uploading
	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(env.Env("DO_S3_BUCKET", "")),
		Key:    aws.String(tempFileName),
		// ACL:                  aws.String("public-read"), // could be private if you want it to be access by only authorized users
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}

	return tempFileName, err
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	maxSize := int64(1024000 * 25) // allow only 1MB of file size *25

	err := r.ParseMultipartForm(maxSize)
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Image too large. Max Size: %v", maxSize)
		return
	}

	file, fileHeader, err := r.FormFile("profile_picture")
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Could not get uploaded file")
		return
	}
	defer file.Close()

	// create an AWS session which can be
	// reused if we're uploading many files
	s, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(
			env.Env("DO_S3_KEY", ""),
			env.Env("DO_S3_SECRET", ""),
			"",
		),
		Endpoint: aws.String(env.Env("DO_S3_ENDPOINT", "https://nyc3.digitaloceanspaces.com")),
		Region:   aws.String(env.Env("DO_S3_REGION", "us-east-1")),
	})
	if err != nil {
		fmt.Fprintf(w, "Could not upload file")
	}

	fileName, err := UploadFileToS3(s, file, fileHeader)
	if err != nil {
		fmt.Fprintf(w, "Could not upload file")
	}

	fmt.Fprintf(w, "Image uploaded successfully: %v", fileName)
}
