package main

// https://stackoverflow.com/questions/34177137/stream-file-upload-to-aws-s3-using-go
// https://medium.com/spankie/upload-images-to-aws-s3-bucket-in-a-golang-web-application-2612bea70dd8
// https://docs.digitalocean.com/products/spaces/resources/s3-sdk-examples/
// https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
// https://codepen.io/PerfectIsShit/pen/zogMXP -- progress bar done on a frontend
// https://github.com/aws/aws-sdk-go/commit/50ba1dfe47983b15b160b66c730a3b93d2961f8e -- progress to stdout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/tomekwlod/booking/core"
	"github.com/tomekwlod/booking/internal/s3"
	us "github.com/tomekwlod/booking/store/user"
	"github.com/tomekwlod/utils/env"
	"golang.org/x/crypto/bcrypt"
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

type signupCredentials struct {
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}
type signupCredentialsError struct {
	Error signupCredentials `json:"errors"`
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

	var creds signupCredentials

	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)

	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		log.Printf("[server/signupHandler] structure of the body is wrong: %+v", creds)

		writeError(w, errBadRequest)

		return
	}

	e := &signupCredentialsError{}

	// some validation need to happen here!!!!
	if creds.Email == "" {

		log.Printf("[server/signupHandler] email cannot be empty: %+v", creds)

		e.Error.Email = "This field cannot be empty"
	}
	if creds.Password == "" {

		log.Printf("[server/signupHandler] password cannot be empty: %+v", creds)

		e.Error.Password = "This field cannot be empty"
	}

	if (signupCredentialsError{}) != *e {
		writeJSON(w, e, http.StatusForbidden)
		return
	}

	// email validation here
	// password validation here

	// Salt and hash the password using the bcrypt algorithm
	// The second argument is the cost of hashing, which we arbitrarily set as 8 (this value can be more or less, depending on the computing power you wish to utilize)
	hash, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	if err != nil {

		log.Printf("[server/signupHandler] problem with hashing the given password: %s", err.Error())
		writeError(w, errInternalServer)
		return
	}

	ctx := context.Background()

	user := core.User{
		Email:    creds.Email,
		Password: string(hash),
	}

	err = dbConn.Transact(func(tx *sqlx.Tx) (err error) {

		userStore := us.New(dbConn)

		//
		// check if user is not already in db
		//

		err = userStore.Create(ctx, tx, &user)
		if err != nil {
			return err
		}

		return
	})

	if err != nil {
		log.Printf("[server/signupHandler] error while writing to a database: %s", err.Error())
		writeError(w, errInternalServer)
		return
	}

	writeJSON(w, user, http.StatusOK)
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
