package utils

import (
	"log"

	"github.com/nicolassutter/scyd/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDatabase() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(EnsureDbPath()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Reduce log verbosity
	})
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return err
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(&models.Download{})
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}
