package models

import (
	"time"

	"gorm.io/gorm"
)

type DownloadState string

const (
	DownloadStatePending  DownloadState = "pending"
	DownloadStateProgress DownloadState = "progress"
	DownloadStateSuccess  DownloadState = "success"
	DownloadStateError    DownloadState = "error"
)

type Download struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	URL          string         `gorm:"not null" json:"url"`
	State        DownloadState  `gorm:"default:pending" json:"state"`
	ErrorMessage string         `gorm:"default:''" json:"error_message"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
