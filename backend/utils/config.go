package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type User struct {
	PasswordHash string `yaml:"password_hash"`
}

// hook_name: command
type Hooks struct {
	OnError            string `yaml:"on_error"`
	OnDownloadComplete string `yaml:"on_download_complete"`
}

type config struct {
	DownloadDir string `yaml:"download_dir"`
	OutputDir   string `yaml:"output_dir"`
	PublicDir   string `yaml:"public_dir"`
	// Automatically sort downloads after each download completes
	SortAfterDownload bool `yaml:"sort_after_download"`
	// Users for authentication
	Users map[string]User `yaml:"users"`
	Hooks Hooks           `yaml:"hooks"`
}

func EnsureDbPath() string {
	DBPath := "./config/scyd.db"

	// Create database parent dir if it doesn't exist
	dbPath := filepath.Clean(DBPath)
	dbDir := filepath.Dir(dbPath)
	err := os.MkdirAll(dbDir, os.ModePerm)

	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	return DBPath
}

func newConfig() *config {
	config := &config{
		DownloadDir:       "/downloads",
		OutputDir:         "/output",
		SortAfterDownload: true,
		Users:             make(map[string]User),
		Hooks:             Hooks{},
		PublicDir:         "/public",
	}
	return config
}

var UserConfig = newConfig()

func IsDevelopment() bool {
	return os.Getenv("GO_ENV") == "development" || os.Getenv("GO_ENV") == ""
}

func ReadUserConfigFile() (*config, error) {
	file, err := os.Open("./config/config.yaml")

	if err != nil {
		fmt.Println("Could not read config file, using defaults.")
	} else {
		defer file.Close()

		decoder := yaml.NewDecoder(file)
		err = decoder.Decode(UserConfig)

		if err != nil {
			fmt.Println("Error reading config file, using defaults.")
		} else {
			fmt.Println("Config file found!")
		}
	}

	// ensure the download dir exists
	err = os.MkdirAll(UserConfig.DownloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating %s dir: %s", err, UserConfig.DownloadDir)
	}

	// ensure the output dir exists
	err = os.MkdirAll(UserConfig.OutputDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Error creating %s dir: %s", err, UserConfig.OutputDir)
	}

	return UserConfig, nil
}
