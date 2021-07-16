package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	app "github.com/w-k-s/simple-budget-tracker/application"
)

var (
	configFilePath string
	awsAccessKey string
	awsSecretKey string
	awsRegion string
	config *app.Config
)

func init(){
	const (
		configFileUsage         = `Path to configFile. Must start with 'file://' (if file is in local filesystem) or 's3://' (if file is hosted on s3)`
		awsAccessKeyUsage	 	= "AWS Access key; used to download config file. Only required if config file is hosted on s3"
		awsSecretKeyUsage		= "AWS Secret key; used to download config file. Only required if config file is hosted on s3"
		awsRegionUsage			= "AWS Region; used to download config file. Only required if config file is hosted on s3"
	)
	flag.StringVar(&configFilePath, "f", "", configFileUsage)
	flag.StringVar(&awsAccessKey, "aws_access_key", "", awsAccessKeyUsage)
	flag.StringVar(&awsSecretKey, "aws_secret_key", "", awsSecretKeyUsage)
	flag.StringVar(&awsRegion, "aws_region", "", awsRegionUsage)

	var err error
	if config,err = app.LoadConfig(configFilePath, awsAccessKey, awsSecretKey, awsRegion); err == nil{
		log.Fatalf("failed to load config file. Reason: %s", err)
	}
}

func main() {
	app.MustRunMigrations("postgres", config.Database().ConnectionString(), config.Database().MigrationDirectory())

	s := &http.Server{
		Addr:           config.Server().ListenAddress(),
		ReadTimeout:    config.Server().ReadTimeout(),
		WriteTimeout:   config.Server().WriteTimeout(),
		MaxHeaderBytes: config.Server().MaxHeaderBytes(),
	}
	log.Fatal(s.ListenAndServe())	
}
