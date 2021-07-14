package application

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// -- SETUP

var configFileContents string =
`
[server]
port = 8080

[database]
username = "jack.torrence"
password = "password"
name     = "overlook"
host     = "localhost"
port     = 5432
sslmode  = "disable"
`

func createTestConfigFile(content string, uri string) error{
	path := strings.Replace(uri, "file://", "", 1)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("Failed to create path '%s'. Reason: %w", path, err)
	}
	if err := ioutil.WriteFile(path, []byte(content), 0777); err != nil {
		return fmt.Errorf("Failed to write test config file to '%s'. Reason: %w", path, err)
	}
	return nil
}

func (suite *ConfigTestSuite) SetupTest() {
	if err := createTestConfigFile(configFileContents, DefaultConfigFilePath()); err != nil{
		log.Fatalf("Failed to create test config file. Reason: %s", err)
	}
}

// -- TEARDOWN

func (suite *ConfigTestSuite) TearDownTest() {
	path := strings.Replace(DefaultConfigFilePath(), "file://", "", 1)
	_ = os.Remove(path)
}

// -- SUITE

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_loadingConfig_THEN_configsLoadedFromDefaultPath() {
	// WHEN
	config, err := LoadConfig("", "", "")

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

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_configFileDoesNotExistAtDefaultPath_THEN_errorIsReturned() {
	// GIVEN
	path := strings.Replace(DefaultConfigFilePath(), "file://", "", 1)
	_ = os.Remove(path)

	// WHEN
	config, err := LoadConfig("", "", "")

	// THEN
	assert.Nil(suite.T(), config)
	assert.NotNil(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf("failed to read config file from local path '%s'. Reason: open %s: no such file or directory", path, path), err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesNotExistAtProvidedPath_THEN_errorIsReturned() {
	// GIVEN
	uri := "file://" + filepath.Join("/.budget", "test.d", "config.toml")

	// WHEN
	config, err := LoadConfig(uri, "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to read config file from local path '/.budget/test.d/config.toml'. Reason: open /.budget/test.d/config.toml: no such file or directory", err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesExistAtProvidedPath_THEN_configsParsedCorrectly() {
	// GIVEN
	assert.Nil(suite.T(), createTestConfigFile(configFileContents, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("", "", "")

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

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileIsEmpty_THEN_errorIsReturned() {
	// GIVEN
	assert.Nil(suite.T(), createTestConfigFile("", DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("", "", "")

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

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsProvided_WHEN_configFileDoesNotContainValidToml_THEN_errorIsReturned() {
	// GIVEN
	invalidToml := `{
		"database":{
			"port":8080
		}
	}`
	assert.Nil(suite.T(), createTestConfigFile(invalidToml, DefaultConfigFilePath()))

	// WHEN
	config, err := LoadConfig("", "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "failed to parse config file. Reason: toml: invalid character at start of key: {", err.Error())
}

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotPrefixedWithFileOrS3Protocol_WHEN_configFileIsLoaded_THEN_errorIsReturned() {
	// GIVEN
	uri := "http://" + filepath.Join("/.budget", "test.d", "config.toml")

	// WHEN
	config, err := LoadConfig(uri, "", "")

	// THEN
	assert.NotNil(suite.T(), err)
	assert.Nil(suite.T(), config)
	assert.Equal(suite.T(), "Config file must start with file:// or s3://", err.Error())
}
