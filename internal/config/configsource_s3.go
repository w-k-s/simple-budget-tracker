package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type s3ConfigSource struct {
	awsAccessKey string
	awsSecretKey string
	awsRegion    string
}

func newS3ConfigSource(accessKey, secretKey, region string) (*s3ConfigSource, error) {
	return &s3ConfigSource{
		awsAccessKey: accessKey,
		awsSecretKey: secretKey,
		awsRegion:    region,
	}, nil
}

type S3Uri struct {
	BucketName string
	Key        string
}

func ParseS3Uri(s3Uri string) (*S3Uri, error) {
	uriWithoutProtcol := strings.Replace(s3Uri, "s3://", "", 1)
	parts := strings.Split(uriWithoutProtcol, "/")
	if len(parts) < 2 || len(parts[1]) == 0 {
		return nil, fmt.Errorf("failed to parse S3 URI. Not enough parts to extract bucket name and object key")
	}
	return &S3Uri{
		BucketName: parts[0],
		Key:        strings.Join(parts[1:], "/"),
	}, nil
}

func (s S3Uri) String() string {
	return fmt.Sprintf("s3://%s/%s", s.BucketName, s.Key)
}

func (s3Config s3ConfigSource) Load(s3UriString string) (*Config, error) {
	var (
		sess *session.Session
		err  error
	)
	if len(s3Config.awsAccessKey) == 0 || len(s3Config.awsSecretKey) == 0 {
		sess, err = session.NewSession()
	} else {
		sess, err = session.NewSessionWithOptions(session.Options{
			Config: aws.Config{
				Region:      aws.String(s3Config.awsRegion),
				Credentials: credentials.NewStaticCredentials(s3Config.awsAccessKey, s3Config.awsSecretKey, ""),
			},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session. Reason: %w", err)
	}

	downloader := s3manager.NewDownloader(sess)

	// Create a file to write the S3 Object contents to.
	tempRoot := strings.Replace(defaultTempDirectoryPath(), "file://", "", 1)
	path := filepath.Join(tempRoot, "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, fmt.Errorf("failed to create temporary directory %q. Reason: %w", path, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file %q, %w", path, err)
	}

	// Write the contents of S3 Object to the file
	var s3Uri *S3Uri
	if s3Uri, err = ParseS3Uri(s3UriString); err != nil {
		return nil, fmt.Errorf("failed to parse s3 Uri: %q. Reason: %w", s3UriString, err)
	}

	if _, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(s3Uri.BucketName),
		Key:    aws.String(s3Uri.Key),
	}); err != nil {
		message := err
		if awsErr, ok := err.(awserr.Error); ok {
			message = fmt.Errorf("[%s]: %s", awsErr.Code(), awsErr.Message())
		}
		return nil, fmt.Errorf("failed to download file from s3 uri %q. Reason: %w", s3Uri, message)
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from local path '%s'. Reason: %w", path, err)
	}
	return readToml(bytes)
}
