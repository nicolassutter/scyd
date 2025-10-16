package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// removes or replaces characters that are invalid in file/directory names
// Windows forbidden characters: < > : " / \ | ? *
// Also removes leading/trailing spaces and dots which can cause issues
func sanitizePathComponent(name string) string {
	// Replace forbidden characters with underscore
	replacer := strings.NewReplacer(
		"<", "_",
		">", "_",
		":", "_",
		"\"", "_",
		"/", "_",
		"\\", "_",
		"|", "_",
		"?", "_",
		"*", "_",
	)
	sanitized := replacer.Replace(name)

	// Trim leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")

	return sanitized
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

		filePath := filepath.Join(utils.UserConfig.DownloadDir, file.Name())
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

		artist = sanitizePathComponent(artist)

		if artist == "" {
			artist = "Unknown Artist"
		}

		album := sanitizePathComponent(metadata.Album())

		// If album is empty, filepath.Join will skip it, creating: /output/Artist/file.mp3
		// If album exists, it creates: /output/Artist/Album/file.mp3
		newDir := filepath.Join(utils.UserConfig.OutputDir, artist, album)

		err = os.MkdirAll(newDir, os.ModePerm)

		if err != nil {
			log.Printf("Failed to create directory %s: %v", newDir, err)
			filesWithErrors = append(filesWithErrors, filePath)
			continue
		}

		newFilePath := filepath.Join(newDir, file.Name())

		// copy the file to the new location
		err = copyFile(filePath, newFilePath)

		if err != nil {
			log.Printf("Failed to copy file %s to %s: %v", filePath, newFilePath, err)
			filesWithErrors = append(filesWithErrors, filePath)
			continue
		}

		// remove the original file
		err = os.Remove(filePath)

		if err != nil {
			log.Printf("Failed to remove original file %s: %v", filePath, err)
			// Note: file was copied successfully, so we still count it as moved
			// but we could optionally clean up the copy if removal fails
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
