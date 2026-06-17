//go:build unit

package web_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/go-angular-spa-template/internal/web"
)

func TestSPA_EmbedPresent(t *testing.T) {
	t.Parallel()
	fsys, err := web.SPA()
	require.NoError(t, err)

	idx, err := fs.ReadFile(fsys, "index.html")
	require.NoError(t, err)
	assert.NotEmpty(t, idx)
}

func TestSPA_Register(t *testing.T) {
	t.Parallel()
	e := echo.New()
	dist := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>shell</html>")},
		"main.js":    &fstest.MapFile{Data: []byte("console.log(1)")},
	}
	require.NoError(t, web.Register(e, dist))

	// Real asset is served with its content type.
	recAsset := httptest.NewRecorder()
	e.ServeHTTP(recAsset, httptest.NewRequest(http.MethodGet, "/main.js", nil))
	assert.Equal(t, http.StatusOK, recAsset.Code)
	assert.Contains(t, recAsset.Body.String(), "console.log")

	// Unknown client-side route falls back to index.html.
	recRoute := httptest.NewRecorder()
	e.ServeHTTP(recRoute, httptest.NewRequest(http.MethodGet, "/tasks/123", nil))
	assert.Equal(t, http.StatusOK, recRoute.Code)
	assert.Contains(t, recRoute.Body.String(), "shell")

	// Root serves index.html.
	recRoot := httptest.NewRecorder()
	e.ServeHTTP(recRoot, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, recRoot.Code)
	assert.Contains(t, recRoot.Body.String(), "shell")
}
