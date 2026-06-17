//go:build unit

package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zercle/go-angular-spa-template/internal/web"
)

func TestSPA_EmbedAccessible(t *testing.T) {
	t.Parallel()
	// The embedded FS is always resolvable; whether index.html is present
	// depends on whether the frontend has been built.
	_, err := web.SPA()
	require.NoError(t, err)
}

func TestSPA_Register_NoIndex_FallsBack(t *testing.T) {
	t.Parallel()
	e := echo.New()
	// Empty dist (frontend not built) → Register must not error and must serve
	// the built-in fallback page.
	require.NoError(t, web.Register(e, fstest.MapFS{}))

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Frontend not built")
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
}
