// Package rest is the HTTP delivery layer built on Fiber v3. It translates
// between HTTP and the use-case layer and serves the embedded Angular SPA.
package rest

import "github.com/gofiber/fiber/v3"

// ok writes a success envelope: {"data": ...}.
func ok(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(fiber.Map{"data": data})
}

// fail writes an error envelope: {"error": "..."}.
func fail(c fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}
