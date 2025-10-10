package handlers

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/shlex"
	"github.com/nicolassutter/scyd/utils"
)

type DownloadResponse struct {
	Body DownloadResponseBody
}
type DownloadResponseBody struct {
	Message         string   `json:"message"`
	DownloadedFiles []string `json:"downloaded_files,omitempty"`
}

func getFileNamesInDownloadDir() ([]string, error) {
	files, err := os.ReadDir(utils.UserConfig.DownloadDir)

	if err != nil {
		return nil, err
	}

	fileNames := []string{}

	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}

func DownloadHandler(ctx context.Context, input *struct {
	Body struct {
		Url       string `required:"true" json:"url"`
		YtDlpArgs string `required:"false" example:"--arg arg_value --second-arg --third-arg" doc:"Pass additional args to yt-dlp" json:"yt_dlp_args"`
	}
}) (*DownloadResponse, error) {
	isDevelopment := utils.IsDevelopment()
	isYoutube := strings.Contains(input.Body.Url, "youtube.com") || strings.Contains(input.Body.Url, "youtu.be")

	allFilesInDownloadDir, err := getFileNamesInDownloadDir()

	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to read download directory: " + err.Error())
	}

	// helper function to get newly downloaded files by comparing files in download dir before and after download
	getNewlyDownloadedFiles := func() []string {
		currentFiles, err := getFileNamesInDownloadDir()

		if err != nil {
			return []string{}
		}

		fileNames := []string{}

		// compare current files with files before download to get newly downloaded files
		for _, fileName := range currentFiles {
			if !slices.Contains(allFilesInDownloadDir, fileName) {
				fileNames = append(fileNames, fileName)
			}
		}

		return fileNames
	}

	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to read download directory: " + err.Error())
	}

	responseBody := &DownloadResponseBody{}

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

		out, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Command error output: %s\n", out)
			return nil, huma.Error500InternalServerError("Failed to execute command: " + err.Error())
		}

		responseBody = &DownloadResponseBody{
			Message:         "Successfully downloaded with yt-dlp!",
			DownloadedFiles: getNewlyDownloadedFiles(),
		}
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

		cmd := exec.Command(command[0], command[1:]...)

		println("Executing command:", strings.Join(cmd.Args, " "))

		out, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("Command error output: %s\n", out)
			return nil, huma.Error500InternalServerError("Failed to execute command: " + err.Error())
		}

		// delete cover.jpg if one has been downloaded
		coverPath := utils.UserConfig.DownloadDir + "/cover.jpg"
		if _, err := os.Stat(coverPath); err == nil {
			err := os.Remove(coverPath)

			if err != nil {
				println("Failed to delete cover.jpg:", err.Error())
			}
		}

		responseBody = &DownloadResponseBody{
			Message:         "Successfully downloaded with streamrip!",
			DownloadedFiles: getNewlyDownloadedFiles(),
		}
	}

	fmt.Printf("Downloaded %s to %s\n", input.Body.Url, utils.UserConfig.DownloadDir)

	if utils.UserConfig.SortAfterDownload {
		_, err := SortDownloadsDirectory()

		if err != nil {
			println("Failed to sort downloads after download:", err.Error())
		}
	}

	return &DownloadResponse{
		Body: *responseBody,
	}, nil
}
