package config

import (
	"log"
	"os"
	"path/filepath"
)

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