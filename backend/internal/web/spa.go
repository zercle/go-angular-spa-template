package web

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v5"
)

// fallbackHTML is served when the Angular build is not embedded yet (e.g. a
// fresh clone where `task build` has not run). The real index.html replaces it
// once the frontend is built.
const fallbackHTML = `<!doctype html>
<html lang="en">
  <head><meta charset="utf-8" /><title>go-angular-spa-template</title></head>
  <body>
    <p>Frontend not built yet. Run <code>task build</code> (or <code>task dev</code>).</p>
  </body>
</html>`

// Register mounts the embedded Angular SPA on the Echo server. Existing files
// (JS/CSS/assets) are served with their content type; any unmatched path falls
// back to index.html so the Angular router can handle client-side routes and
// deep-link refreshes. Explicit routes (e.g. /api/v1, /healthz) take priority
// because Echo's router ranks static and param routes above the "/*" wildcard.
func Register(e *echo.Echo, dist fs.FS) error {
	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		index = []byte(fallbackHTML) // frontend not built yet
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
