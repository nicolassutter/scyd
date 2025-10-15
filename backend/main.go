package main

import (
	"context"
	"log"

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

	// Initialize database
	err := utils.InitDatabase()
	if err != nil {
		log.Fatalf("Error initializing database: %s", err)
	}

	fiberApp := fiber.New()

	if utils.IsDevelopment() {
		fiberApp.Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:3001",
			AllowCredentials: true,
		}))
	}

	api := humafiber.New(fiberApp, huma.DefaultConfig("scyd REST API", "1.0.0"))

	// Register global middleware that allows to access the Fiber context from Huma handlers
	api.UseMiddleware(utils.StoreFiberCtx)

	api_v1 := huma.NewGroup(api, "/api/v1")

	// Auth routes (public) - now using Huma
	huma.Post(api_v1, "/auth/login", handlers.LoginHandler)
	huma.Post(api_v1, "/auth/logout", handlers.LogoutHandler)
	huma.Get(api_v1, "/auth/status", handlers.AuthStatusHandler)

	// health check route
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

	// Protected routes (require authentication)
	// Apply auth middleware to the Huma API routes by using Fiber middleware
	// Huma is protected as well because it's mounted on the Fiber app
	fiberApiV1 := fiberApp.Group("/api/v1", handlers.AuthMiddleware())

	// Download routes (protected)
	huma.Post(api_v1, "/download", handlers.DownloadHandler)
	huma.Delete(api_v1, "/download/{id}", handlers.DeleteDownloadHandler)
	huma.Post(api_v1, "/sort-downloads", handlers.SortDownloadsHandler)
	huma.Get(api_v1, "/downloads", handlers.GetDownloadsHandler)

	// Setup WebSocket for real-time download updates
	handlers.SetupDownloadWebSocket(&fiberApiV1)

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
