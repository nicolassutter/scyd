package services

import (
	"github.com/nicolassutter/scyd/models"
	"github.com/nicolassutter/scyd/utils"
)

type DownloadService struct{}

func NewDownloadService() *DownloadService {
	return &DownloadService{}
}

func (ds *DownloadService) CreateDownload(url string) (*models.Download, error) {
	download := &models.Download{
		URL:   url,
		State: models.DownloadStatePending,
	}

	result := utils.DB.Create(download)
	if result.Error != nil {
		return nil, result.Error
	}

	return download, nil
}

func (ds *DownloadService) DeleteDownload(id uint) error {
	result := utils.DB.Delete(&models.Download{}, id)
	return result.Error
}

func (ds *DownloadService) UpdateDownloadState(id uint, state models.DownloadState, errorMessage string) error {
	result := utils.DB.Model(&models.Download{}).Where("id = ?", id).Updates(map[string]interface{}{
		"state":         state,
		"error_message": errorMessage,
	})

	return result.Error
}

func (ds *DownloadService) GetDownload(id uint) (*models.Download, error) {
	var download models.Download
	result := utils.DB.First(&download, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &download, nil
}

func (ds *DownloadService) GetAllDownloads() ([]models.Download, error) {
	var downloads []models.Download
	result := utils.DB.Order("created_at DESC").Find(&downloads)
	if result.Error != nil {
		return nil, result.Error
	}

	return downloads, nil
}
