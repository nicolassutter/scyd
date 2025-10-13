package handlers

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/shlex"
	"github.com/google/uuid"
	"github.com/nicolassutter/scyd/utils"
)

type DownloadResponse struct {
	Body DownloadResponseBody
}
type DownloadResponseBody struct {
	Message string `json:"message"`
	TaskID  string `json:"task_id"`
}

var taskStatus = make(map[string]chan string) // map[TaskID] -> Channel for sending log lines
var mu sync.Mutex                             // Mutex to safely access the map

func startDownloadTask(taskId string, outputCh chan string, cmd *exec.Cmd) huma.StatusError {
	defer func() {
		// Clean up the map and close the channel when done
		mu.Lock()
		delete(taskStatus, taskId)
		close(outputCh)
		mu.Unlock()
	}()

	// Create pipes to capture both stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed to get command stdout:", err.Error())
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return huma.Error500InternalServerError("Failed to get command stdout: " + err.Error())
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Failed to get command stderr:", err.Error())
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return huma.Error500InternalServerError("Failed to get command stderr: " + err.Error())
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Failed to start command:", err.Error())
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
		return huma.Error500InternalServerError("Failed to start command: " + err.Error())
	}

	// 3. Read and Stream both stdout and stderr
	// Use goroutines to read both streams concurrently
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("STDOUT: %s\n", line)

			select {
			case outputCh <- line:
			case <-context.Background().Done():
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("STDERR: %s\n", line)

			select {
			case outputCh <- line:
			case <-context.Background().Done():
				return
			}
		}
	}()

	// Wait for the command to finish and close the connection
	cmd.Wait()

	if utils.UserConfig.SortAfterDownload {
		fmt.Printf("Sorting downloads directory %s\n", utils.UserConfig.DownloadDir)

		_, err := SortDownloadsDirectory()

		if err != nil {
			utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
			fmt.Println("Failed to sort downloads after download:", err.Error())
		}
	}

	// delete cover.jpg if one has been downloaded
	coverPath := utils.UserConfig.DownloadDir + "/cover.jpg"
	if _, err := os.Stat(coverPath); err == nil {
		err := os.Remove(coverPath)

		if err != nil {
			println("Failed to delete cover.jpg:", err.Error())
		}
	}

	return nil
}

type DownloadStdoutNewLineEvent struct {
	Line string `json:"line" example:"[download] Downloading video 1 of 1"`
}

// RawDownloadStreamHandler handles SSE using raw Fiber without Huma's SSE wrapper
func RawDownloadStreamHandler(c *fiber.Ctx) error {
	taskID := c.Params("task_id")

	// Set SSE headers
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	mu.Lock()
	ch, exists := taskStatus[taskID]
	mu.Unlock()

	if !exists {
		// Send not found event, empty data
		c.WriteString(fmt.Sprintf("event: download_not_found\ndata: %s\n\n", "{}"))
		return nil
	}

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		for line := range ch {
			event := DownloadStdoutNewLineEvent{Line: line}
			data, _ := json.Marshal(event)

			// Write SSE format
			fmt.Fprintf(w, "event: new_line\ndata: %s\n\n", string(data))

			// Flush the writer
			err := w.Flush()
			if err != nil {
				fmt.Printf("SSE flush error (continuing): %v\n", err)
			}
		}

		// Send completion event, empty data
		fmt.Fprintf(w, "event: download_success\ndata: %s\n\n", "{}")
		utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnDownloadComplete)
		w.Flush()
	})

	return nil
}

func DownloadHandler(ctx context.Context, input *struct {
	Body struct {
		Url       string `required:"true" json:"url"`
		YtDlpArgs string `required:"false" example:"--arg arg_value --second-arg --third-arg" doc:"Pass additional args to yt-dlp" json:"yt_dlp_args"`
	}
}) (*DownloadResponse, error) {
	// 1. Generate a unique Task ID
	taskId := uuid.New().String()

	// 2. Create a channel for this task's output
	taskChannel := make(chan string)
	mu.Lock()
	taskStatus[taskId] = taskChannel
	mu.Unlock()

	isDevelopment := utils.IsDevelopment()
	isYoutube := strings.Contains(input.Body.Url, "youtube.com") || strings.Contains(input.Body.Url, "youtu.be")

	cmd := &exec.Cmd{}

	if isYoutube {
		additionalArgs, err := shlex.Split(input.Body.YtDlpArgs)

		if err != nil {
			errorMessage := fmt.Sprintf(
				"Failed to parse additional yt-dlp args '%s': %s\n", input.Body.YtDlpArgs, err.Error(),
			)
			utils.ExecuteCommandBg(utils.UserConfig.Hooks.OnError)
			return nil, huma.Error400BadRequest(errorMessage)
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
	go startDownloadTask(taskId, taskChannel, cmd)

	fmt.Printf("Download started for: %s to %s\n", input.Body.Url, utils.UserConfig.DownloadDir)

	return &DownloadResponse{
		Body: DownloadResponseBody{
			Message: "Download started",
			TaskID:  taskId,
			// DownloadedFiles: newlyDownloadedFiles,
		},
	}, nil
}
