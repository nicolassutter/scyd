package main

import (
	"context"
	"slices"
	"strings"

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
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(slices.Compact(utils.UserConfig.AllowOrigins), ","),
		AllowCredentials: true,
	}))

	api := humafiber.New(fiberApp, huma.DefaultConfig("scyd REST API", "1.0.0"))

	api_v1 := huma.NewGroup(api, "/api/v1")

	huma.Post(api_v1, "/download", handlers.DownloadHandler)
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
	fiberApp.Listen(":3000")
}
