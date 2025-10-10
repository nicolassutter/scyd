package utils

import (
	"github.com/dhowden/tag"
	"log"
	"os"
)

func GetMetadataFromFile(audioFilePath string) (tag.Metadata, error) {
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
