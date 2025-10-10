package handlers

import (
	"context"
	"log"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/nicolassutter/scyd/utils"
)

type SortDownloadsResponseBody struct {
	MovedFiles      []string `json:"moved_files"`
	FilesWithErrors []string `json:"files_with_errors"`
}
type SortDownloadsResponse struct {
	Body SortDownloadsResponseBody
}

// sort every audio file in the downloads dir into artist/album folders, then move them to the output dir
func SortDownloadsDirectory() (*SortDownloadsResponse, error) {
	files, err := os.ReadDir(utils.UserConfig.DownloadDir)

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

		filePath := utils.UserConfig.DownloadDir + "/" + file.Name()
		metadata, err := utils.GetMetadataFromFile(filePath)

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

		newDir := utils.UserConfig.OutputDir + "/" + artist + "/" + album

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
}

func SortDownloadsHandler(c context.Context, input *struct{}) (*SortDownloadsResponse, error) {
	return SortDownloadsDirectory()
}
