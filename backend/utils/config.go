package utils

import (
	"fmt"
	"log"
	"os"

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
	// Automatically sort downloads after each download completes
	SortAfterDownload bool `yaml:"sort_after_download"`
	// Users for authentication
	Users map[string]User `yaml:"users"`
	Hooks Hooks           `yaml:"hooks"`
}

func newConfig() *config {
	config := &config{
		DownloadDir:       "/downloads",
		OutputDir:         "/output",
		SortAfterDownload: true,
		Users:             make(map[string]User),
		Hooks:             Hooks{},
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
