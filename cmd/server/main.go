package main

import (
	"flag"
	"log"
	"net/http"

	cfg "github.com/w-k-s/simple-budget-tracker/internal/config"
	db "github.com/w-k-s/simple-budget-tracker/internal/persistence"
	app "github.com/w-k-s/simple-budget-tracker/internal/server"
)

var (
	configFilePath string
	awsAccessKey   string
	awsSecretKey   string
	awsRegion      string
)

func init() {
	const (
		configFileUsage   = `Path to configFile. Must start with 'file://' (if file is in local filesystem) or 's3://' (if file is hosted on s3)`
		awsAccessKeyUsage = "AWS Access key; used to download config file. Only required if config file is hosted on s3"
		awsSecretKeyUsage = "AWS Secret key; used to download config file. Only required if config file is hosted on s3"
		awsRegionUsage    = "AWS Region; used to download config file. Only required if config file is hosted on s3"
	)
	flag.StringVar(&configFilePath, "file", "", configFileUsage)
	flag.StringVar(&awsAccessKey, "aws_access_key", "", awsAccessKeyUsage)
	flag.StringVar(&awsSecretKey, "aws_secret_key", "", awsSecretKeyUsage)
	flag.StringVar(&awsRegion, "aws_region", "", awsRegionUsage)
}

func main() {
	// LoadConfig must be called in the main function and not in the init function because
	// the init function is called in tests but the config file does not exist.
	// This results in a panic.
	flag.Parse()

	err := cfg.ConfigureLogging(); 
	if err != nil {
		log.Fatalf("failed to configure logging. Reason: %s", err)
	}

	config, err := cfg.LoadConfig(configFilePath, awsAccessKey, awsSecretKey, awsRegion); 
	if err != nil {
		log.Fatalf("failed to load config file. Reason: %s", err)
	}

	handler, err := app.Init(config); 
	if err != nil {
		log.Fatalf("failed to init application. Reason: %s", err)
	}

	db.MustRunMigrations(
		config.Database(),
	)

	s := &http.Server{
		Addr:           config.Server().ListenAddress(),
		Handler:        handler.Router(),
		ReadTimeout:    config.Server().ReadTimeout(),
		WriteTimeout:   config.Server().WriteTimeout(),
		MaxHeaderBytes: config.Server().MaxHeaderBytes(),
	}
	log.Fatal(s.ListenAndServe())
}
