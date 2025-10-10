package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/dhowden/tag"
	"github.com/gofiber/fiber/v2"
	"github.com/google/shlex"
	"gopkg.in/yaml.v3"
)

type DownloadRequest struct {
	Url       string `json:"url"`
	YtDlpArgs string `json:"yt_dlp_args"`
}

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
}

func readConfigFile() (*Config, error) {
	config := new(Config)
	file, err := os.Open("./config/config.yaml")

	if err != nil {
		return config, err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		fmt.Println("Error reading config file, using defaults.")
	}

	fmt.Println("Config file found!")

	defaults := Config{
		DownloadDir: "/downloads",
		OutputDir:   "/output",
	}

	if config.DownloadDir == "" {
		config.DownloadDir = defaults.DownloadDir
	}
	if config.OutputDir == "" {
		config.OutputDir = defaults.OutputDir
	}

	// ensure the download dir exists
	err = os.MkdirAll(config.DownloadDir, os.ModePerm)
	if err != nil {
		return config, err
	}

	// ensure the output dir exists
	err = os.MkdirAll(config.OutputDir, os.ModePerm)
	if err != nil {
		return config, err
	}

	return config, nil
}

func main() {
	config, err := readConfigFile()

	if err != nil {
		log.Printf("Failed to read config file: %v", err)
	}

	app := fiber.New()
	appv1 := app.Group("/api/v1")

	appv1.Post("/download", func(c *fiber.Ctx) error {
		req := new(DownloadRequest)

		if err := c.BodyParser(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Cannot parse JSON",
			})
		}

		isYoutube := strings.Contains(req.Url, "youtube.com") || strings.Contains(req.Url, "youtu.be")
		// isSoundCloud := strings.Contains(req.Url, "soundcloud.com")

		if isYoutube {
			additionalArgs, err := shlex.Split(req.YtDlpArgs)

			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":  "Failed to parse yt-dlp arguments",
					"detail": err.Error(),
				})
			}

			command := []string{
				"docker",
				"run",
				"--rm",
				"-v",
				config.DownloadDir + ":/downloads",
				"scyd",
				"yt-dlp",
				"-o",
				"/downloads/%(title)s.%(ext)s",
				"--extract-audio",
				"--audio-format",
				"mp3",
				"--audio-quality",
				"0",
				"--embed-thumbnail",
				"--add-metadata",
			}

			// add additional args from request
			command = append(command, additionalArgs...)
			// finally add the url
			command = append(command, req.Url)

			cmd := exec.Command(command[0], command[1:]...)

			println("Executing command:", strings.Join(cmd.Args, " "))

			_, err = cmd.Output()

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":  "Failed to execute command",
					"detail": err.Error(),
					"stderr": err.(*exec.ExitError).Stderr,
				})
			}

			return c.JSON(fiber.Map{
				"message": "Successfully downloaded from Youtube with: " + req.Url,
			})
		} else {
			cmd := exec.Command(
				"docker",
				"run",
				"--rm",
				"-v",
				config.DownloadDir+":/downloads",
				"scyd",
				"rip",
				"-f",
				"/downloads",
				"url",
				req.Url,
			)

			println("Executing command:", strings.Join(cmd.Args, " "))

			_, err := cmd.Output()

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":  "Failed to execute command",
					"detail": err.Error(),
					"stderr": err.(*exec.ExitError).Stderr,
				})
			}

			return c.JSON(fiber.Map{
				"message": "Successfully downloaded from SoundCloud with: " + req.Url,
			})
		}
	})

	appv1.Get("/metadata", func(c *fiber.Ctx) error {
		audioFilePath := c.Query("file")

		if audioFilePath == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "File query parameter is required",
			})
		}

		metadata, err := getMetadataFromFile(audioFilePath)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to get metadata",
				"detail": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"title":  metadata.Title(),
			"artist": metadata.Artist(),
			"album":  metadata.Album(),
			"year":   metadata.Year(),
		})
	})

	// sort every audio file in the downloads dir into artist/album folders, then move them to the output dir
	appv1.Post("/sort-downloads", func(c *fiber.Ctx) error {
		files, err := os.ReadDir(config.DownloadDir)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":  "Failed to read download directory",
				"detail": err.Error(),
			})
		}

		movedFiles := []string{}

		for _, file := range files {
			if file.IsDir() {
				continue
			}

			filePath := config.DownloadDir + "/" + file.Name()
			metadata, err := getMetadataFromFile(filePath)

			if err != nil {
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
				continue
			}

			newFilePath := newDir + "/" + file.Name()

			// move the file
			err = os.Rename(filePath, newFilePath)

			if err != nil {
				log.Printf("Failed to move file %s to %s: %v", filePath, newFilePath, err)
				continue
			}

			movedFiles = append(movedFiles, newFilePath)
		}

		return c.JSON(fiber.Map{
			"moved_files": movedFiles,
		})
	})

	app.Listen(":3000")
}
