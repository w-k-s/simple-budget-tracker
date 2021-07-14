package application

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	server ServerConfig
	db     DBConfig
}

func (c Config) Server() ServerConfig {
	return c.server
}

func (c Config) Database() DBConfig {
	return c.db
}

type ServerConfig struct {
	port int
}

func (s ServerConfig) Port() int {
	return s.port
}

type DBConfig struct {
	username     string
	password     string
	host         string
	port         int
	name         string
	sslMode      string
	migrationDir string
}

func (d DBConfig) Username() string {
	return d.username
}

func (d DBConfig) Password() string {
	return d.password
}

func (d DBConfig) Host() string {
	return d.host
}

func (d DBConfig) Port() int {
	return d.port
}

func (d DBConfig) Name() string {
	return d.name
}

func (d DBConfig) SslMode() string {
	return d.sslMode
}

func (d DBConfig) MigrationDirectory() string {
	if len(d.migrationDir) == 0 {
		return DefaultMigrationsDirectoryPath()
	}
	return d.migrationDir
}

func (d DBConfig) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host(),
		d.Port(),
		d.Username(),
		d.Password(),
		d.Name(),
		d.SslMode(),
	)
}

func readToml(bytes []byte) (*Config, error) {
	var mutableConfig struct {
		Server struct {
			Port int
		}
		Database struct {
			Username     string
			Password     string
			Host         string
			Port         int
			Name         string
			SSLMode      string
			MigrationDir string
		}
	}

	err := toml.Unmarshal(bytes, &mutableConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file. Reason: %w", err)
	}

	config := &Config{
		server: ServerConfig{
			port: mutableConfig.Server.Port,
		},
		db: DBConfig{
			username:     mutableConfig.Database.Username,
			password:     mutableConfig.Database.Password,
			host:         mutableConfig.Database.Host,
			port:         mutableConfig.Database.Port,
			name:         mutableConfig.Database.Name,
			sslMode:      mutableConfig.Database.SSLMode,
			migrationDir: mutableConfig.Database.MigrationDir,
		},
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

type localConfigSource struct{}

func (l localConfigSource) Load(configFilePath string) (*Config, error) {
	localPath := strings.Replace(configFilePath, "file://", "", 1)
	bytes, err := ioutil.ReadFile(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file from local path '%s'. Reason: %w", localPath, err)
	}
	return readToml(bytes)
}

type s3ConfigSource struct {
	awsAccessKey string
	awsSecretKey string
}

func newS3ConfigSource(accessKey, secretKey string) (*s3ConfigSource, error) {
	if len(accessKey) == 0 && len(secretKey) == 0{
		log.Print(
			`AWS Access Key and AWS Secret Access Key not provided. 
			Will try to fetch these credentials from Environment Variables / ~/.aws/credentials.txt or that the server has appopriate AWS role.`,
		)
	}
	return &s3ConfigSource{
		awsAccessKey: accessKey,
		awsSecretKey: secretKey,
	}, nil
}

func (s3 s3ConfigSource) Load(objectPath string) (*Config, error) {
	return nil, fmt.Errorf("to be implemented")
}

func DefaultConfigFilePath() string {
	return "file://" + filepath.Join(MustUserHomeDir(), ".budget", "config.toml")
}

func DefaultMigrationsDirectoryPath() string {
	return filepath.Join(MustUserHomeDir(), ".budget", "migrations.d")
}

func MustUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to access user's home directory")
	}
	return homeDir
}

func LoadConfig(configFilePath, awsAccessKey, awsSecretKey string) (*Config, error) {
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
		if s3Source, err = newS3ConfigSource(awsAccessKey, awsSecretKey); err != nil {
			return nil, err
		}
		return s3Source.Load(configFilePath)
	}

	return nil, fmt.Errorf("Config file must start with file:// or s3://")
}
