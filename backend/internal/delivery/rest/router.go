package rest

import (
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/cerzzlive/gofiber-angular-spa/backend/internal/usecase"
)

// NewRouter builds the Fiber app: middleware, the versioned /api/v1 routes, and
// the embedded Angular SPA (static assets + client-side-routing fallback).
func NewRouter(svc *usecase.TaskService, spa fs.FS) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "gofiber-angular-spa",
	})

	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api/v1")
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	NewTaskHandler(svc).Register(api)

	registerSPA(app, spa)
	return app
}

// registerSPA serves the embedded Angular build. Existing files (JS/CSS/assets)
// are served with their content type; any unmatched path returns index.html so
// the Angular router can handle client-side routes and deep-link refreshes.
func registerSPA(app *fiber.App, spa fs.FS) {
	index, _ := fs.ReadFile(spa, "index.html")

	app.Get("/*", func(c fiber.Ctx) error {
		p := strings.TrimPrefix(c.Path(), "/")
		if p == "" {
			return c.Type("html").Send(index)
		}

		data, err := fs.ReadFile(spa, p)
		if err != nil {
			// Not a real asset → SPA shell for client-side routing.
			return c.Type("html").Send(index)
		}
		if ct := mime.TypeByExtension(filepath.Ext(p)); ct != "" {
			c.Set(fiber.HeaderContentType, ct)
		}
		return c.Send(data)
	})
}
