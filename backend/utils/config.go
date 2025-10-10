package utils

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type config struct {
	DownloadDir string `yaml:"download_dir"`
	OutputDir   string `yaml:"output_dir"`
	// Automatically sort downloads after each download completes
	SortAfterDownload bool `yaml:"sort_after_download"`
}

func newConfig() *config {
	config := &config{
		DownloadDir:       "/downloads",
		OutputDir:         "/output",
		SortAfterDownload: true,
	}
	return config
}

var UserConfig = newConfig()

func ReadUserConfigFile() (*config, error) {
	file, err := os.Open("./config/config.yaml")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(UserConfig)

	if err != nil {
		fmt.Println("Error reading config file, using defaults.")
	} else {
		fmt.Println("Config file found!")
	}

	// ensure the download dir exists
	err = os.MkdirAll(UserConfig.DownloadDir, os.ModePerm)
	if err != nil {
		return UserConfig, err
	}

	// ensure the output dir exists
	err = os.MkdirAll(UserConfig.OutputDir, os.ModePerm)
	if err != nil {
		return UserConfig, err
	}

	return UserConfig, nil
}
