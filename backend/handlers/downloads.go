package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/shlex"
	"github.com/nicolassutter/scyd/models"
	"github.com/nicolassutter/scyd/services"
	"github.com/nicolassutter/scyd/utils"
)

type DownloadResponse struct {
	Body DownloadResponseBody
}

type DownloadResponseBody struct {
	Message    string `json:"message"`
	DownloadID uint   `json:"download_id"`
}

func broadcastDownloadMessage(msg DownloadMessage) {
	msgBytes, err := json.Marshal(msg)

	if err != nil {
		fmt.Println("Failed to marshal download message:", err)
		return
	}

	for _, client := range clients {
		if client != nil {
			client.Emit(msgBytes)
		}
	}
}

type DownloadEvent string

const (
	DownloadEventStart    DownloadEvent = "start"
	DownloadEventProgress DownloadEvent = "progress"
	DownloadEventError    DownloadEvent = "error"
	DownloadEventSuccess  DownloadEvent = "success"
)

type DownloadMessage struct {
	Event      DownloadEvent `json:"event"`
	DownloadID uint          `json:"download_id"`
	Data       string        `json:"data"`
}

// key: connection uuid
var clients = make(map[string]*socketio.Websocket)

// WebSocket handler for download connections
func SetupDownloadWebSocket(router *fiber.Router) {
	socketio.On(socketio.EventDisconnect, func(payload *socketio.EventPayload) {
		delete(clients, payload.Kws.UUID)
	})

	// require websocket upgrade to access this route
	(*router).Use("/ws/download", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	(*router).Get("/ws/download", socketio.New(func(kws *socketio.Websocket) {
		clients[kws.UUID] = kws
	}))
}

func startDownloadTaskWS(downloadID uint, cmd *exec.Cmd) {
	downloadService := services.NewDownloadService()
	var errorMessage string

	defer func() {
		// Update download state based on command result
		if cmd.ProcessState != nil && cmd.ProcessState.Success() {
			downloadService.UpdateDownloadState(downloadID, models.DownloadStateSuccess, "")
			utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnDownloadComplete)
			// Broadcast success message
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventSuccess,
				DownloadID: downloadID,
				Data:       "Download completed successfully",
			})
		} else {
			if errorMessage == "" {
				errorMessage = "Download failed"
			}
			downloadService.UpdateDownloadState(downloadID, models.DownloadStateError, errorMessage)
			utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)

			// Broadcast error message
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventError,
				DownloadID: downloadID,
				Data:       "Download failed",
			})
		}

		// Post-process: sort downloads if configured
		if utils.UserConfig.SortAfterDownload {
			fmt.Printf("Sorting downloads directory %s\n", utils.UserConfig.DownloadDir)
			_, err := SortDownloadsDirectory()
			if err != nil {
				utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
				fmt.Println("Failed to sort downloads after download:", err.Error())
			}
		}

		// Delete cover.jpg if one has been downloaded
		coverPath := utils.UserConfig.DownloadDir + "/cover.jpg"
		if _, err := os.Stat(coverPath); err == nil {
			err := os.Remove(coverPath)
			if err != nil {
				fmt.Printf("Failed to delete cover.jpg: %s\n", err.Error())
			}
		}
	}()

	// Create pipes to capture both stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errorMessage = "Failed to get command stdout: " + err.Error()
		fmt.Println(errorMessage)
		downloadService.UpdateDownloadState(downloadID, models.DownloadStateError, errorMessage)
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		errorMessage = "Failed to get command stderr: " + err.Error()
		fmt.Println(errorMessage)
		downloadService.UpdateDownloadState(downloadID, models.DownloadStateError, errorMessage)
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		errorMessage = "Failed to start download command: " + err.Error()
		fmt.Println(errorMessage)
		downloadService.UpdateDownloadState(downloadID, models.DownloadStateError, errorMessage)
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return
	}

	// Update download state to progress
	downloadService.UpdateDownloadState(downloadID, models.DownloadStateProgress, "")

	// Broadcast start message
	broadcastDownloadMessage(DownloadMessage{
		Event:      DownloadEventStart,
		DownloadID: downloadID,
		Data:       "Download started",
	})

	// Read stdout in a goroutine and send updates to WebSocket clients
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("STDOUT: %s\n", line)

			// Broadcast progress update
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventProgress,
				DownloadID: downloadID,
				Data:       line,
			})
		}
	}()

	// Read stderr in another goroutine and capture errors
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("STDERR: %s\n", line)

			// Capture error messages from stderr
			if strings.Contains(strings.ToLower(line), "error") ||
				strings.Contains(strings.ToLower(line), "failed") ||
				strings.Contains(strings.ToLower(line), "cannot") {
				errorMessage = line
			}

			// Broadcast error output
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventError,
				DownloadID: downloadID,
				Data:       line,
			})
		}
	}()

	// Wait for the command to finish
	cmd.Wait()
}

// DownloadHandler handles download requests with WebSocket streaming
func DownloadHandler(ctx context.Context, input *struct {
	Body struct {
		Url       string `required:"true" json:"url"`
		YtDlpArgs string `required:"false" example:"--arg arg_value --second-arg --third-arg" doc:"Pass additional args to yt-dlp" json:"yt_dlp_args"`
	}
}) (*DownloadResponse, error) {
	// 1. Create download record in database
	downloadService := services.NewDownloadService()
	download, err := downloadService.CreateDownload(input.Body.Url)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to create download record")
	}

	isYoutube := strings.Contains(input.Body.Url, "youtube.com") || strings.Contains(input.Body.Url, "youtu.be")

	var cmd *exec.Cmd

	isDevelopment := utils.IsDevelopment()

	if isYoutube {
		additionalArgs := []string{}
		if input.Body.YtDlpArgs != "" {
			additionalArgs, err = shlex.Split(input.Body.YtDlpArgs)
			if err != nil {
				return nil, huma.Error400BadRequest(fmt.Sprintf(
					"Failed to parse additional yt-dlp args '%s': %s\n", input.Body.YtDlpArgs, err.Error(),
				))
			}
		}

		devDockerPrefix := []string{
			"docker",
			"run",
			"--rm",
			"-v",
			utils.UserConfig.DownloadDir + ":" + utils.UserConfig.DownloadDir,
			"scyd",
		}

		ytDlpBaseCommand := []string{
			"yt-dlp",
			"-o",
			utils.UserConfig.DownloadDir + "/%(title)s.%(ext)s",
			"--extract-audio",
			"--audio-format",
			"mp3",
			"--audio-quality",
			"0",
			"--embed-thumbnail",
			"--add-metadata",
			"--progress",  // Force progress output
			"--newline",   // Force newlines in output
			"--no-colors", // Disable colors for cleaner parsing
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

		cmd = exec.Command(command[0], command[1:]...)
	} else {
		// handle other platforms like Soundcloud with streamrip

		devDockerPrefix := []string{
			"docker",
			"run",
			"--rm",
			"-v",
			utils.UserConfig.DownloadDir + ":" + utils.UserConfig.DownloadDir,
			"scyd",
		}

		baseCmd := []string{
			"rip",
			"-f",
			utils.UserConfig.DownloadDir,
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

		cmd = exec.Command(command[0], command[1:]...)
	}

	// Start the download task in a separate goroutine so we don't block
	go startDownloadTaskWS(download.ID, cmd)

	fmt.Printf("Download started for: %s to %s\n", input.Body.Url, utils.UserConfig.DownloadDir)

	return &DownloadResponse{
		Body: DownloadResponseBody{
			Message:    "Download started",
			DownloadID: download.ID,
		},
	}, nil
}

type GetDownloadsResponse struct {
	Body GetDownloadsResponseBody
}

type GetDownloadsResponseBody struct {
	Downloads []models.Download `json:"downloads"`
}

// GetDownloadsHandler returns all downloads from the database
func GetDownloadsHandler(ctx context.Context, input *struct{}) (*GetDownloadsResponse, error) {
	downloadService := services.NewDownloadService()
	downloads, err := downloadService.GetAllDownloads()
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get downloads: " + err.Error())
	}

	return &GetDownloadsResponse{
		Body: GetDownloadsResponseBody{
			Downloads: downloads,
		},
	}, nil
}
