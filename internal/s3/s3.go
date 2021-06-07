package s3

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
)

type digitaloceanClient struct {
	endpoint, region, key, secret, bucket, dir string
}

func New(endpoint, region, key, secret, bucket, dir string) *digitaloceanClient {
	return &digitaloceanClient{
		endpoint,
		region,
		key,
		secret,
		bucket,
		dir,
	}
}

func (do *digitaloceanClient) Upload(mp *multipart.Part) (string, error) {

	awsConfig := &aws.Config{
		Endpoint: aws.String(do.endpoint),
		Region:   aws.String(do.region),

		Credentials: credentials.NewStaticCredentials(
			do.key,
			do.secret,
			"",
		),
	}

	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession(awsConfig))

	destinationFile, err := do.uploadFileToS3(sess, mp)

	if err != nil {
		return "", errors.New(fmt.Sprintf("Cannot upload file, %s", err.Error()))
	}

	return destinationFile, nil
}

// uploadFileToS3 saves a file to aws bucket and returns the url to // the file and an error if there's any
func (do *digitaloceanClient) uploadFileToS3(sess *session.Session, file *multipart.Part) (string, error) {

	// Create an uploader with the session and custom options
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // The minimum/default allowed part size is 5MB
		u.Concurrency = 2            // default is 5
	})

	fileext := filepath.Ext(file.FileName())
	filename := file.FileName()[0 : len(file.FileName())-len(fileext)]

	// create a unique file name for the file
	tempFileName := path.Join(
		do.dir,
		fmt.Sprintf("%s_%s%s", filename, uuid.New().String(), filepath.Ext(file.FileName())),
	)

	// Upload the file to S3.
	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(do.bucket),
		Key:    aws.String(tempFileName),
		Body:   file,
	})

	// in case it fails to upload
	if err != nil {
		return "", err
	}

	return result.Location, nil
}
