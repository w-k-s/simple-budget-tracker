package config

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	server ServerConfig
	db     DBConfig
	gpt    GptConfig
}

func NewConfig(
	serverConfig ServerConfig, 
	dbConfig DBConfig,
	gptConfig GptConfig,
) (*Config, error) {
	config := &Config{
		server: serverConfig,
		db:     dbConfig,
		gpt: 	gptConfig,
	}

	errors := validate.Validate(
		&validators.IntIsGreaterThan{Name: "Server Port", Field: int(config.server.port), Compared: 1023, Message: "Server port must be at least 1023"},
		&validators.StringLengthInRange{Name: "Database Username", Field: config.db.username, Min: 1, Max: 0, Message: "Database username is required"},
		&validators.StringLengthInRange{Name: "Database Password", Field: config.db.password, Min: 1, Max: 0, Message: "Database password is required"},
		&validators.StringLengthInRange{Name: "Database Host", Field: config.db.host, Min: 1, Max: 0, Message: "Database host is required"},
		&validators.IntIsGreaterThan{Name: "Database Port", Field: int(config.db.port), Compared: 0, Message: "Database port is required"},
		&validators.StringLengthInRange{Name: "Database Name", Field: config.db.host, Min: 1, Max: 0, Message: "Database name is required"},
		&validators.StringInclusion{Name: "Database SSL Mode", Field: config.db.sslMode, List: []string{"disable", "require", "verify-ca", "verify-full"}, Message: "Database SSL Mode is required"},
		&validators.StringLengthInRange{Name: "Migration Directory", Field: config.db.host, Min: 1, Max: 0, Message: "Migration Directory path is required"},
	)

	if errors.HasAny() {
		return nil, errors
	}

	return config, nil
}

func (c Config) Server() ServerConfig {
	return c.server
}

func (c Config) Database() DBConfig {
	return c.db
}

func (c Config) Gpt() GptConfig{
	return c.gpt
}

func readToml(bytes []byte) (*Config, error) {
	var mutableConfig struct {
		Server struct {
			Port                int
			WriteTimeoutSeconds int64 `toml:"write_timeout"`
			ReadTimeoutSeconds  int64 `toml:"read_timeout"`
			MaxHeaderBytes      int   `toml:"max_header_bytes"`
		}
		Database struct {
			Username     string
			Password     string
			Host         string
			Port         int
			Name         string
			SSLMode      string
			MigrationDir string `toml:"migration_dir"`
		}
		Gpt struct {
			ApiKey       string `toml:"api_key"`
		}
	}

	err := toml.Unmarshal(bytes, &mutableConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file. Reason: %w", err)
	}

	return NewConfig(
		ServerConfig{
			port:           mutableConfig.Server.Port,
			readTimeout:    time.Duration(mutableConfig.Server.ReadTimeoutSeconds) * time.Second,
			writeTimeout:   time.Duration(mutableConfig.Server.WriteTimeoutSeconds) * time.Second,
			maxHeaderBytes: mutableConfig.Server.MaxHeaderBytes,
		},
		DBConfig{
			username:     mutableConfig.Database.Username,
			password:     mutableConfig.Database.Password,
			host:         mutableConfig.Database.Host,
			port:         mutableConfig.Database.Port,
			name:         mutableConfig.Database.Name,
			sslMode:      mutableConfig.Database.SSLMode,
			migrationDir: mutableConfig.Database.MigrationDir,
		},
		GptConfig{
			apiKey: mutableConfig.Gpt.ApiKey,	
		},
	)
}

type localConfigSource struct{}

func (l localConfigSource) Load(configFilePath string) (*Config, error) {
	localPath := strings.Replace(configFilePath, "file://", "", 1)
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from local path '%s'. Reason: %w", localPath, err)
	}
	return readToml(bytes)
}

func LoadConfig(configFilePath, awsAccessKey, awsSecretKey, awsRegion string) (*Config, error) {
	if len(configFilePath) == 0 {
		configFilePath = DefaultConfigFilePath()
	}

	if strings.HasPrefix(configFilePath, "file://") {
		return localConfigSource{}.Load(configFilePath)
	}

	if strings.HasPrefix(configFilePath, "s3://") {
		var (
			s3Source *s3ConfigSource
			err      error
		)
		if s3Source, err = newS3ConfigSource(awsAccessKey, awsSecretKey, awsRegion); err != nil {
			return nil, err
		}
		return s3Source.Load(configFilePath)
	}

	return nil, fmt.Errorf("Config file must start with file:// or s3://")
}
