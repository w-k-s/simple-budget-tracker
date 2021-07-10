package application

import (
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

var configFileContents string

func (suite *ConfigTestSuite) SetupSuite() {

	configFileContents =
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
	path := strings.Replace(DefaultConfigFilePath(), "file://", "", 1)
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		log.Fatalf("Failed to create path '%s'. Reason: %s", path, err)
	}
	if err := ioutil.WriteFile(path, []byte(configFileContents), 0777); err != nil {
		log.Fatalf("Failed to write test config file to '%s'. Reason: %s", path, err)
	}
}

// -- TEARDOWN

func (suite *ConfigTestSuite) TearDownSuite() {
	path := strings.Replace(DefaultConfigFilePath(), "file://", "", 1)
	err := os.Remove(path)
	if err != nil {
		log.Fatalf("Failed to delete test config file at '%s'. Reason: %s", path, err)
	}
}

// -- SUITE

func (suite *ConfigTestSuite) Test_GIVEN_configFilePathIsNotProvided_WHEN_loadingConfig_THEN_configsLoadedFromDefaultPath() {
	// WHEN
	config, err := LoadConfig("", "", "")

	// THEN
	assert.Nil(suite.T(), err)
	assert.NotNil(suite.T(), config)
	assert.Equal(suite.T(), 8080, config.Server().Port())
}
