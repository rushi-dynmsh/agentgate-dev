// Package apidocs serves AgentGate's governance REST API (internal/govapi)
// as interactive, browsable documentation.
//
// It does not describe the runtime MCP enforcement boundary (internal/authz)
// — that's a gRPC Envoy v3 ext_authz service, not a request/response REST
// API, and isn't representable in OpenAPI/Swagger.
//
// Both routes this package registers are deliberately unauthenticated: they
// serve only the API's static shape (the OpenAPI document and a Swagger UI
// page), never live data. Swagger UI's own "Try it out" feature still
// requires a real admin credential for every actual call to a govapi route,
// exactly as a curl request would. This is a documentation surface, not a
// new trust boundary.
package apidocs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var specYAML []byte

// uiHTML loads Swagger UI from a CDN (jsdelivr) rather than vendoring it —
// keeps this a zero-new-Go-dependency addition. Requires the browser
// viewing /docs to have internet access; the API itself does not.
const uiHTML = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>AgentGate Governance API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis],
      });
    };
  </script>
</body>
</html>
`

// RegisterRoutes registers the API documentation endpoints on mux:
//   - GET /openapi.yaml — the raw OpenAPI 3.0 spec
//   - GET /docs         — interactive Swagger UI, pointed at the spec above
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(specYAML)
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(uiHTML))
	})
}
