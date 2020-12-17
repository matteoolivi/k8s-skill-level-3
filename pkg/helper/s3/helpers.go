package s3

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// TODO: Harmonize this signature with that of DownloadImage: the uploader should be a parameter
// that gets passed in at every function invocation rather than something that gets created in the
// function body.
func UploadImage(imageName string, image io.Reader) (string, error) {
	uploader, err := GetUploader()
	if err != nil {
		return "", err
	}
	log.Printf("Successfully connected to S3 to upload image %s.", imageName)

	imageURL, err := upload(uploader, imageName, image)
	if err != nil {
		return "", err
	}
	log.Printf("Successfully uploaded image %s to S3 bucket.", imageName)

	return imageURL, nil
}

func DownloadImage(imageName string, downloader *s3manager.Downloader, dest os.File) error {
	sizeBytes, err := downloader.Download(&dest,
		&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv(s3BucketEnvVar)),
			Key:    aws.String(imageName),
		})
	if err != nil {
		return fmt.Errorf("failed to download image %s: %s",
			imageName,
			err)
	}
	log.Printf("Successfully downloaded image %s (~= %d KiBs) to file %s",
		imageName,
		sizeBytes/1024,
		dest.Name())

	return nil
}

func GetUploader() (*s3manager.Uploader, error) {
	// TODO: For efficiency (and maybe other reasons as well) we might re-use the same config and
	// session across multiple requests, rather than building everything from scratch at every
	// request. Or, stop using session and use a stateless client instead, if that's a thing.
	sess, err := session.NewSession(newAwsCfg())
	if err != nil {
		return nil, fmt.Errorf("failed to create aws session: %s", err)
	}
	log.Printf("Successfully created aws session.")

	return s3manager.NewUploader(sess), nil
}

func GetDownloader() (*s3manager.Downloader, error) {
	// TODO: if possible, refactor to share this code with GetUploader.
	sess, err := session.NewSession(newAwsCfg())
	if err != nil {
		return nil, fmt.Errorf("failed to create aws session: %s", err)
	}
	log.Printf("Successfully created aws session.")

	return s3manager.NewDownloader(sess), nil
}

const s3BucketEnvVar = "S3_BUCKET"

func upload(s3 *s3manager.Uploader, imageName string, image io.Reader) (string, error) {
	output, err := s3.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv(s3BucketEnvVar)),
		Body:   image,
		Key:    aws.String(imageName),
	})
	if err != nil {
		return "", err
	}

	return output.Location, nil
}

// TODO: Read credentials from config file rather than env vars (or, support both possibilities).
const (
	awsRegionEnvVar = "AWS_REGION"
	awsIDEnvVar     = "AWS_ACCESS_KEY"
	awsSecretEnvVar = "AWS_SECRET_KEY"
)

func newAwsCfg() *aws.Config {
	return aws.
		NewConfig().
		WithRegion(os.Getenv(awsRegionEnvVar)).
		WithCredentials(credentials.NewEnvCredentials())
}
