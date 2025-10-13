package utils

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humafiber"
	"github.com/gofiber/fiber/v2"
)

func StoreFiberCtx(ctx huma.Context, next func(huma.Context)) {
	fiberCtx := humafiber.Unwrap(ctx)
	ctx = huma.WithValue(ctx, "fiberCtx", fiberCtx)
	next(ctx)
}

func GetFiberCtx(ctx context.Context) *fiber.Ctx {
	return ctx.Value("fiberCtx").(*fiber.Ctx)
}
