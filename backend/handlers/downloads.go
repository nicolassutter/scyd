package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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

	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

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
var clientsMutex sync.RWMutex

type DownloadManager struct {
	downloads map[uint]context.CancelFunc
	mu        sync.RWMutex
}

var downloadManager = &DownloadManager{
	downloads: make(map[uint]context.CancelFunc),
}

// stores a cancel function for a download
func (dm *DownloadManager) StoreDownload(downloadID uint, cancel context.CancelFunc) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.downloads[downloadID] = cancel
}

// cancels a download and removes it from the map
// returns `true` if a download was cancelled or `false` if not found
func (dm *DownloadManager) CancelDownload(downloadID uint) bool {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if cancel, exists := dm.downloads[downloadID]; exists {
		cancel()
		delete(dm.downloads, downloadID)
		return true
	}
	return false
}

// removes a download from the map (for cleanup after completion)
func (dm *DownloadManager) RemoveDownload(downloadID uint) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	delete(dm.downloads, downloadID)
}

// WebSocket handler for download connections
func SetupDownloadWebSocket(router *fiber.Router) {
	socketio.On(socketio.EventDisconnect, func(payload *socketio.EventPayload) {
		clientsMutex.Lock()
		delete(clients, payload.Kws.UUID)
		clientsMutex.Unlock()
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
		clientsMutex.Lock()
		clients[kws.UUID] = kws
		clientsMutex.Unlock()
	}))
}

func startDownloadTaskWS(downloadID uint, commandArgs []string) {
	downloadService := services.NewDownloadService()
	var errorMessage string

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Store cancel function for potential cancellation
	downloadManager.StoreDownload(downloadID, cancel)

	// Ensure download is removed from map when done
	defer downloadManager.RemoveDownload(downloadID)

	defer func() {
		// Update download state based on command result

		// the context was cancelled
		if ctx.Err() == context.Canceled {
			downloadService.UpdateDownloadState(downloadID, models.DownloadStateError, "Download cancelled")
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventError,
				DownloadID: downloadID,
				Data:       "Download cancelled",
			})
		} else if errorMessage == "" { // success
			downloadService.UpdateDownloadState(downloadID, models.DownloadStateSuccess, "")
			utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnDownloadComplete)
			// Broadcast success message
			broadcastDownloadMessage(DownloadMessage{
				Event:      DownloadEventSuccess,
				DownloadID: downloadID,
				Data:       "Download completed successfully",
			})
		} else { // error occurred
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

	// create a command instance with our context for automatic cancellation
	cmd := exec.CommandContext(ctx, commandArgs[0], commandArgs[1:]...)

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

	// Wait for the command to finish (or be cancelled)
	err = cmd.Wait()
	// Check if context was cancelled otherwise capture error
	if err != nil && ctx.Err() != context.Canceled {
		errorMessage = "Command failed: " + err.Error()
	}
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

	isDevelopment := utils.IsDevelopment()

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
		// output template: optional_artist dash_if_artist_not_empty title - [extractor] [track_id].ext
		filepath.Join(utils.UserConfig.DownloadDir, "%(artist)s%(artist& - )s%(title)s - [%(extractor)s] [%(track_id,id)s].%(ext)s"),
		"--extract-audio",
		"--audio-format",
		"mp3",
		"--audio-quality",
		"0",
		"--embed-thumbnail",
		"--embed-metadata",
		"--windows-filenames",
		"--progress",  // Force progress output
		"--newline",   // Force newlines in output
		"--no-colors", // Disable colors for cleaner parsing
	}

	downloadCommandArgs := []string{}

	// run inside a docker container in development
	if isDevelopment {
		downloadCommandArgs = append(downloadCommandArgs, devDockerPrefix...)
		downloadCommandArgs = append(downloadCommandArgs, ytDlpBaseCommand...)
	} else {
		// in production just run yt-dlp directly
		downloadCommandArgs = append(downloadCommandArgs, ytDlpBaseCommand...)
	}

	// add additional args from request
	downloadCommandArgs = append(downloadCommandArgs, additionalArgs...)

	// finally add the url
	downloadCommandArgs = append(downloadCommandArgs, input.Body.Url)

	// Start the download task in a separate goroutine so we don't block
	go startDownloadTaskWS(download.ID, downloadCommandArgs)

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

func DeleteDownloadHandler(ctx context.Context, input *struct {
	ID uint `required:"true" path:"id"`
}) (*struct{}, error) {
	downloadService := services.NewDownloadService()

	err := downloadService.DeleteDownload(input.ID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to delete download: " + err.Error())
	}

	return nil, nil
}

func CancelDownloadHandler(ctx context.Context, input *struct {
	ID uint `required:"true" path:"id"`
}) (*struct{}, error) {
	if downloadManager.CancelDownload(input.ID) {
		return nil, nil
	}

	return nil, huma.Error409Conflict("No active download with the given ID")
}
