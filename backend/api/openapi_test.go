//go:build unit

package api_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/zercle/go-angular-spa-template/api"
)

// TestOpenAPISpec_DocumentsTaskRoutes is a drift guard: every route the tasks
// HTTP handler registers must be documented in openapi.yaml. Structural OpenAPI
// validity is enforced separately by the Spectral lint job in CI; here we only
// parse the spec and assert route coverage so the hand-maintained doc can't
// silently fall out of sync with features/tasks/handler/http.Register.
func TestOpenAPISpec_DocumentsTaskRoutes(t *testing.T) {
	var doc struct {
		OpenAPI string                    `yaml:"openapi"`
		Paths   map[string]map[string]any `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(api.OpenAPISpec, &doc), "openapi.yaml must parse")
	require.NotEmpty(t, doc.OpenAPI, "openapi version must be set")

	want := map[string][]string{
		"/api/v1/tasks":      {"get", "post"},
		"/api/v1/tasks/{id}": {"get", "put", "delete"},
	}
	for path, methods := range want {
		ops, ok := doc.Paths[path]
		require.Truef(t, ok, "openapi.yaml is missing path %q", path)
		for _, m := range methods {
			require.Containsf(t, ops, m, "openapi.yaml is missing %s %s", m, path)
		}
	}
}
