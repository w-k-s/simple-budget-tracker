package application

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigS3TestSuite struct {
	suite.Suite
}

func TestConfigS3TestSuite(t *testing.T) {
	suite.Run(t, new(ConfigS3TestSuite))
}

// -- SETUP

var testS3Location = "s3://com.wks.budget/test/config.toml"

func uploadTestConfigFile(content string, s3UriString string) error {
	path := strings.Replace(DefaultConfigFilePath(), "file://", "", 1)

	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("Failed to create path '%s'. Reason: %w", path, err)
	}
	if err := ioutil.WriteFile(path, []byte(content), 0777); err != nil {
		return fmt.Errorf("Failed to write test config file to '%s'. Reason: %w", path, err)
	}

	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %w", path, err)
	}

	// Upload the file to S3.
	s3Uri, _ := ParseS3Uri(s3UriString)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Uri.BucketName),
		Key:    aws.String(s3Uri.Key),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object %q. Reason: %w", s3Uri, err)
	}
	fmt.Printf("config file uploaded to %q", result.Location)
	return nil
}

func deleteTestConfigFile(s3UriString string) error {
	s3Uri, _ := ParseS3Uri(s3UriString)

	svc := s3.New(session.Must(session.NewSession()))
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3Uri.BucketName),
		Key:    aws.String(s3Uri.Key),
	}

	_, err := svc.DeleteObject(input)
	return err
}

func (suite *ConfigS3TestSuite) SetupTest() {
	if err := uploadTestConfigFile(configFileContents, testS3Location); err != nil {
		log.Fatalf("Failed to upload test config file to bucket %q. Reason: %s", testS3Location, err)
	}
}

// -- TEARDOWN

func (suite *ConfigS3TestSuite) TearDownTest() {
	_ = deleteTestConfigFile(testS3Location)
}

// -- SUITE

func (suite *ConfigS3TestSuite) Test_GIVEN_ValidS3UriString_WHEN_itIsParsed_THEN_bucketNameAndObjectKeyAreExtractedCorrectly() {
	// GIVEN
	uriString := "s3://com.wks.budget/toast/config.yml"

	// WHEN
	s3Uri, err := ParseS3Uri(uriString)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), s3Uri)
	assert.Equal(suite.T(), "com.wks.budget", s3Uri.BucketName)
	assert.Equal(suite.T(), "toast/config.yml", s3Uri.Key)
}

func (suite *ConfigS3TestSuite) Test_GIVEN_ValidS3Uri_WHEN_stringIsCalled_THEN_s3UriStringIsReturned() {
	// GIVEN
	uriString := "s3://com.wks.budget/toast/config.yml"

	// WHEN
	s3Uri, err := ParseS3Uri(uriString)

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), s3Uri)
	assert.Equal(suite.T(), uriString, s3Uri.String())
}

func (suite *ConfigS3TestSuite) Test_GIVEN_s3UriWithoutKey_WHEN_parseIsCalled_THEN_errorIsReturned() {
	// GIVEN
	uriString := "s3://com.wks.budget/"

	// WHEN
	s3Uri, err := ParseS3Uri(uriString)

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), s3Uri)
	assert.Equal(suite.T(), "failed to parse S3 URI. Not enough parts to extract bucket name and object key", err.Error())
}

func (suite *ConfigS3TestSuite) Test_GIVEN_s3UriIsProvided_WHEN_configFileDoesNotExistAtProvidedPath_THEN_errorIsReturned() {
	// GIVEN
	uri := "s3://com.wks.budget/toast/config.yml"

	// WHEN
	config, err := LoadConfig(uri, "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to download file from s3 uri \"s3://com.wks.budget/toast/config.yml\". Reason: [NoSuchKey]: The specified key does not exist.", err.Error())
}

func (suite *ConfigS3TestSuite) Test_GIVEN_s3UriIsProvided_WHEN_configFileDoesExistAtProvidedPath_THEN_configsParsedCorrectly() {
	// GIVEN
	assert.Nil(suite.T(), uploadTestConfigFile(configFileContents, testS3Location))

	// WHEN
	config, err := LoadConfig(testS3Location, "", "", "")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), 8080, config.Server().Port())
	assert.Equal(suite.T(), "jack.torrence", config.Database().Username())
	assert.Equal(suite.T(), "password", config.Database().Password())
	assert.Equal(suite.T(), "overlook", config.Database().Name())
	assert.Equal(suite.T(), 5432, config.Database().Port())
	assert.Equal(suite.T(), "disable", config.Database().SslMode())
	assert.Equal(suite.T(), "host=localhost port=5432 user=jack.torrence password=password dbname=overlook sslmode=disable", config.Database().ConnectionString())
}

func (suite *ConfigS3TestSuite) Test_GIVEN_s3UriIsProvided_WHEN_configFileIsEmpty_THEN_errorIsReturned() {
	// GIVEN
	assert.Nil(suite.T(), uploadTestConfigFile("", testS3Location))

	// WHEN
	config, err := LoadConfig(testS3Location, "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Contains(suite.T(), err.Error(), "Database SSL Mode is required")
	assert.Contains(suite.T(), err.Error(), "Database host is required")
	assert.Contains(suite.T(), err.Error(), "Server port must be at least 1023")
	assert.Contains(suite.T(), err.Error(), "Database username is required")
	assert.Contains(suite.T(), err.Error(), "Migration Directory path is required")
	assert.Contains(suite.T(), err.Error(), "Database password is required")
	assert.Contains(suite.T(), err.Error(), "Database port is required")
	assert.Contains(suite.T(), err.Error(), "Database name is required")
}

func (suite *ConfigS3TestSuite) Test_GIVEN_s3UriIsProvided_WHEN_configFileDoesNotContainValidToml_THEN_errorIsReturned() {
	// GIVEN
	invalidToml := `{
		"database":{
			"port":8080
		}
	}`
	assert.Nil(suite.T(), uploadTestConfigFile(invalidToml, testS3Location))

	// WHEN
	config, err := LoadConfig(testS3Location, "", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to parse config file. Reason: toml: invalid character at start of key: {", err.Error())
}
