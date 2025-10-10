package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/dhowden/tag"
	"github.com/gofiber/fiber/v2"
	"github.com/google/shlex"
	"gopkg.in/yaml.v3"
)

func getMetadataFromFile(audioFilePath string) (tag.Metadata, error) {
	// Check if the file exists
	if _, err := os.Stat(audioFilePath); os.IsNotExist(err) {
		log.Fatalf("Audio file not found: %s", audioFilePath)
	}

	file, err := os.Open(audioFilePath)

	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}

	defer file.Close()

	metadata, err := tag.ReadFrom(file)

	if err != nil {
		return nil, err
	}

	return metadata, nil
}

type Config struct {
	DownloadDir string `yaml:"download_dir"`
	OutputDir   string `yaml:"output_dir"`
	// Automatically sort downloads after each download completes
	SortAfterDownload bool `yaml:"sort_after_download"`
}

func newConfig() *Config {
	config := &Config{
		DownloadDir:       "/downloads",
		OutputDir:         "/output",
		SortAfterDownload: true,
	}
	return config
}

func readConfigFile() (*Config, error) {
	parsedConfig := newConfig()

	file, err := os.Open("./config/config.yaml")

	if err != nil {
		return nil, err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(parsedConfig)

	if err != nil {
		fmt.Println("Error reading config file, using defaults.")
	} else {
		fmt.Println("Config file found!")
	}

	// ensure the download dir exists
	err = os.MkdirAll(parsedConfig.DownloadDir, os.ModePerm)
	if err != nil {
		return parsedConfig, err
	}

	// ensure the output dir exists
	err = os.MkdirAll(parsedConfig.OutputDir, os.ModePerm)
	if err != nil {
		return parsedConfig, err
	}

	return parsedConfig, nil
}

type DownloadResponse struct {
	Status int
	Body   DownloadResponseBody
}
type DownloadResponseBody struct {
	Message string `json:"message"`
}

func main() {
	config, err := readConfigFile()

	// isProduction := os.Getenv("GO_ENV") == "production"
	isDevelopment := os.Getenv("GO_ENV") == "development" || os.Getenv("GO_ENV") == ""

	if err != nil {
		log.Printf("Failed to read config file: %v", err)
	}

	fiberApp := fiber.New()
	api := humafiber.New(fiberApp, huma.DefaultConfig("scyd REST API", "1.0.0"))

	api_v1 := huma.NewGroup(api, "/api/v1")

	huma.Post(api_v1, "/download", func(ctx context.Context, input *struct {
		Body struct {
			Url       string `json:"url"`
			YtDlpArgs string `json:"yt_dlp_args"`
		}
	}) (*DownloadResponse, error) {
		isYoutube := strings.Contains(input.Body.Url, "youtube.com") || strings.Contains(input.Body.Url, "youtu.be")

		if isYoutube {
			additionalArgs, err := shlex.Split(input.Body.YtDlpArgs)

			if err != nil {
				return nil, huma.Error400BadRequest(("Failed to parse yt-dlp args: " + err.Error()))
			}

			devDockerPrefix := []string{
				"docker",
				"run",
				"--rm",
				"-v",
				config.DownloadDir + ":" + config.DownloadDir,
				"scyd",
			}

			ytDlpBaseCommand := []string{
				"yt-dlp",
				"-o",
				config.DownloadDir + "/%(title)s.%(ext)s",
				"--extract-audio",
				"--audio-format",
				"mp3",
				"--audio-quality",
				"0",
				"--embed-thumbnail",
				"--add-metadata",
			}

			command := []string{}

			// run inside a docker container in development
			if isDevelopment {
				command = append(command, devDockerPrefix...)
				command = append(command, ytDlpBaseCommand...)
			} else {
				// in production just run yt-dlp directly
				command = append(command, ytDlpBaseCommand...)
			}

			// add additional args from request
			command = append(command, additionalArgs...)

			// finally add the url
			command = append(command, input.Body.Url)

			cmd := exec.Command(command[0], command[1:]...)

			println("Executing command:", strings.Join(cmd.Args, " "))

			_, err = cmd.Output()

			if err != nil {
				return nil, huma.Error500InternalServerError("Failed to execute command: " + err.Error())
			}

			return &DownloadResponse{
				Body: DownloadResponseBody{
					Message: "Successfully downloaded from Youtube with: " + input.Body.Url,
				},
			}, nil
		} else {
			// handle other platforms like Soundcloud with streamrip

			devDockerPrefix := []string{
				"docker",
				"run",
				"--rm",
				"-v",
				config.DownloadDir + ":" + config.DownloadDir,
				"scyd",
			}

			baseCmd := []string{
				"rip",
				"-f",
				config.DownloadDir,
				"url",
				input.Body.Url,
			}

			command := []string{}

			// run inside a docker container in development
			if isDevelopment {
				command = append(command, devDockerPrefix...)
				command = append(command, baseCmd...)
			} else {
				// in production just run rip directly
				command = append(command, baseCmd...)
			}

			cmd := exec.Command(command[0], command[1:]...)

			println("Executing command:", strings.Join(cmd.Args, " "))

			_, err := cmd.Output()

			if err != nil {
				return nil, huma.Error500InternalServerError("Failed to execute command: " + err.Error())
			}

			return &DownloadResponse{
				Body: DownloadResponseBody{
					Message: "Successfully downloaded from SoundCloud with: " + input.Body.Url,
				},
			}, nil
		}
	})

	type SortDownloadsResponseBody struct {
		MovedFiles      []string `json:"moved_files"`
		FilesWithErrors []string `json:"files_with_errors"`
	}
	type SortDownloadsResponse struct {
		Body SortDownloadsResponseBody
	}

	// sort every audio file in the downloads dir into artist/album folders, then move them to the output dir
	huma.Post(api_v1, "/sort-downloads", func(c context.Context, input *struct{}) (*SortDownloadsResponse, error) {
		files, err := os.ReadDir(config.DownloadDir)

		if err != nil {
			log.Printf("Failed to read download directory: %v", err)
			return nil, huma.Error500InternalServerError("Failed to read download directory")
		}

		movedFiles := []string{}
		filesWithErrors := []string{}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filePath := config.DownloadDir + "/" + file.Name()
			metadata, err := getMetadataFromFile(filePath)

			if err != nil {
				// if we fail to get metadata, skip the file as it might not be an audio file
				log.Printf("Failed to get metadata for file %s: %v", filePath, err)
				continue
			}

			artist := metadata.Artist()
			if artist == "" {
				artist = "Unknown Artist"
			}

			album := metadata.Album()
			if album == "" {
				album = "Unknown Album"
			}

			newDir := config.OutputDir + "/" + artist + "/" + album

			err = os.MkdirAll(newDir, os.ModePerm)

			if err != nil {
				log.Printf("Failed to create directory %s: %v", newDir, err)
				filesWithErrors = append(filesWithErrors, filePath)
				continue
			}

			newFilePath := newDir + "/" + file.Name()

			// move the file
			err = os.Rename(filePath, newFilePath)

			if err != nil {
				log.Printf("Failed to move file %s to %s: %v", filePath, newFilePath, err)
				filesWithErrors = append(filesWithErrors, filePath)
				continue
			}

			movedFiles = append(movedFiles, newFilePath)
		}

		return &SortDownloadsResponse{
			Body: SortDownloadsResponseBody{
				MovedFiles:      movedFiles,
				FilesWithErrors: filesWithErrors,
			},
		}, nil
	})

	fiberApp.Listen(":3000")
}
