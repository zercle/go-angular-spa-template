package web

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
)

// Register mounts the embedded Angular SPA on the Echo server. Existing files
// (JS/CSS/assets) are served with their content type; any unmatched path falls
// back to index.html so the Angular router can handle client-side routes and
// deep-link refreshes. Explicit routes (e.g. /api/v1, /healthz) take priority
// because Echo's router ranks static and param routes above the "/*" wildcard.
func Register(e *echo.Echo, dist fs.FS) error {
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		return fmt.Errorf("read index.html: %w", err)
	}

	e.GET("/*", func(c *echo.Context) error {
		p := strings.TrimPrefix(c.Request().URL.Path, "/")
		if p == "" {
			return c.HTMLBlob(http.StatusOK, index)
		}

		data, readErr := fs.ReadFile(dist, p)
		if readErr != nil {
			// Not a real asset → serve the SPA shell for client-side routing.
			return c.HTMLBlob(http.StatusOK, index)
		}

		ctype := mime.TypeByExtension(path.Ext(p))
		if ctype == "" {
			ctype = http.DetectContentType(data)
		}
		return c.Blob(http.StatusOK, ctype, data)
	})

	return nil
}
