// Package api embeds the OpenAPI specification so the HTTP server can serve it
// (GET /api/openapi.yaml) and render it as interactive docs (GET /api/docs).
// The spec in openapi.yaml is the hand-maintained source of truth for the REST
// API; keep it in sync when you add or change endpoints.
package api

import _ "embed"

// OpenAPISpec is the raw OpenAPI 3.1 document for the REST API.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
