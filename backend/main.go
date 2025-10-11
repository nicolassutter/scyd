package main

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/nicolassutter/scyd/handlers"
	"github.com/nicolassutter/scyd/utils"
)

func main() {
	// sets up the user config before anything else
	utils.ReadUserConfigFile()

	fiberApp := fiber.New()

	if utils.IsDevelopment() {
		fiberApp.Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:3001",
			AllowCredentials: true,
		}))
	}

	api := humafiber.New(fiberApp, huma.DefaultConfig("scyd REST API", "1.0.0"))

	api_v1 := huma.NewGroup(api, "/api/v1")

	huma.Post(api_v1, "/download", handlers.DownloadHandler)
	// Use a raw Fiber handler for SSE to avoid flushing issues
	fiberApp.Get("/api/v1/download/stream/:task_id", handlers.RawDownloadStreamHandler)
	huma.Post(api_v1, "/sort-downloads", handlers.SortDownloadsHandler)

	type statusResponse struct {
		Body *fiber.Map
	}
	huma.Get(api_v1, "/status", func(ctx context.Context, input *struct{}) (*statusResponse, error) {
		return &statusResponse{
			Body: &fiber.Map{
				"status": "ok",
			},
		}, nil
	})

	if !utils.IsDevelopment() {
		fiberApp.Static("/", "./public")
		/**
		 * Serve manifest.webmanifest with correct MIME type
		 */
		fiberApp.Get("/manifest.webmanifest", func(c *fiber.Ctx) error {
			err := c.SendFile("./public/manifest.webmanifest")
			if err != nil {
				return err
			}
			c.Set("Content-Type", "application/manifest+json")
			return nil
		})
	}
	fiberApp.Listen(":3000")
}
