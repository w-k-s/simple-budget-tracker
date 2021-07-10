package main

import (
	"flag"
	"log"
	"os"

	app "github.com/w-k-s/simple-budget-tracker/application"
	"github.com/w-k-s/simple-budget-tracker/migrations"
)

var (
	configFilePath string
	awsAccessKey string
	awsSecretKey string
)

func init(){
	const (
		configFileUsage         = `Path to configFile. Must start with 'file://' (if file is in local filesystem) or 's3://' (if file is hosted on s3)`
		awsAccessKeyUsage	 	= "AWS Access key; used to download config file. Only required if config file is hosted on s3"
		awsSecretKeyUsage		= "AWS Secret key; used to download config file. Only required if config file is hosted on s3"
	)
	flag.StringVar(&configFilePath, "f", "", configFileUsage)
	flag.StringVar(&awsAccessKey, "aws_access_key", "", awsAccessKeyUsage)
	flag.StringVar(&awsSecretKey, "aws_secret_key", "", awsSecretKey)

	var err error
	if _,err = app.LoadConfig(configFilePath, awsAccessKey, awsSecretKey); err == nil{
		log.Fatalf("failed to load config file. Reason: %s", err)
	}
}

func main() {
	migrations.MustRunMigrations("postgres", "postgres://localhost:5432/simple_budget_tracker?sslmode=disable",os.Getenv("MIGRATIONS_DIRECTORY"))
}
