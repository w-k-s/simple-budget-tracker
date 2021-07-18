package application

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	server ServerConfig
	db     DBConfig
}

func NewConfig(serverConfig ServerConfig, dbConfig DBConfig) (*Config, error) {
	config := &Config{
		server: serverConfig,
		db:     dbConfig,
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

type ServerConfig struct {
	port           int
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxHeaderBytes int
}

func (s ServerConfig) Port() int {
	return s.port
}

func (s ServerConfig) MaxHeaderBytes() int {
	if s.maxHeaderBytes <= 0 {
		return 1 << 20 // 1MB
	}
	return s.maxHeaderBytes
}

func (s ServerConfig) ReadTimeout() time.Duration {
	if s.readTimeout == 0 {
		return 10 * time.Second
	}
	return s.readTimeout
}

func (s ServerConfig) WriteTimeout() time.Duration {
	if s.writeTimeout == 0 {
		return 10 * time.Second
	}
	return s.writeTimeout
}

func (s ServerConfig) ListenAddress() string {
	return fmt.Sprintf(":%d", s.port)
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

func (d DBConfig) DriverName() string {
	return "postgres"
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

func DefaultApplicationRootDirectory() string {
	return filepath.Join(mustUserHomeDir(), ".budget")
}

func DefaultConfigFilePath() string {
	return "file://" + filepath.Join(DefaultApplicationRootDirectory(), "config.toml")
}

func DefaultMigrationsDirectoryPath() string {
	return "file://" + filepath.Join(DefaultApplicationRootDirectory(), "migrations.d")
}

func defaultTempDirectoryPath() string {
	return "file://" + filepath.Join(DefaultApplicationRootDirectory(), "temporary.d")
}

func defaultLogsDirectoryPath() string {
	return filepath.Join(DefaultApplicationRootDirectory(), "logs.d")
}

func mustUserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to access user's home directory")
	}
	return homeDir
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

func ConfigureLogging() error{
	var err error
	path := filepath.Join(defaultLogsDirectoryPath(), "server.log")
	if err = os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return fmt.Errorf("failed to create temporary directory %q. Reason: %w", path, err)
	}

	if _, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil{
		return fmt.Errorf("failed to open log file. Reason: %w", err)
	}
	lumberjackLogger := &lumberjack.Logger{
        Filename:   path, 
        MaxSize:    5, // MB
        MaxBackups: 10,
        MaxAge:     30,   // days
        Compress:   true, // disabled by default
    }

	multiWriter := io.MultiWriter(os.Stderr, lumberjackLogger)
	log.SetOutput(multiWriter)

	return nil
}
