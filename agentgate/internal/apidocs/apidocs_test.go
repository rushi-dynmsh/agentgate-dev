package apidocs

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRoutes_ServesSpecAndSwaggerUI(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	specResp, err := http.Get(ts.URL + "/openapi.yaml")
	if err != nil {
		t.Fatalf("GET /openapi.yaml: %v", err)
	}
	defer specResp.Body.Close()
	if specResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /openapi.yaml, got %d", specResp.StatusCode)
	}
	if ct := specResp.Header.Get("Content-Type"); !strings.Contains(ct, "yaml") {
		t.Fatalf("expected a yaml content-type, got %q", ct)
	}
	specBody, err := io.ReadAll(specResp.Body)
	if err != nil {
		t.Fatalf("read spec body: %v", err)
	}
	if !strings.HasPrefix(string(specBody), "openapi: 3.0") {
		t.Fatalf("expected the spec to start with an openapi 3.0 version declaration, got: %.60s", specBody)
	}

	docsResp, err := http.Get(ts.URL + "/docs")
	if err != nil {
		t.Fatalf("GET /docs: %v", err)
	}
	defer docsResp.Body.Close()
	if docsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /docs, got %d", docsResp.StatusCode)
	}
	if ct := docsResp.Header.Get("Content-Type"); !strings.Contains(ct, "html") {
		t.Fatalf("expected an html content-type, got %q", ct)
	}
	docsBody, err := io.ReadAll(docsResp.Body)
	if err != nil {
		t.Fatalf("read docs body: %v", err)
	}
	if !strings.Contains(string(docsBody), "/openapi.yaml") {
		t.Fatalf("expected the docs page to point Swagger UI at /openapi.yaml")
	}
}

// TestOpenAPISpec_CoversEveryGovAPIRoute is a deliberately simple smoke test:
// it string-searches the raw embedded YAML for every route path currently
// registered in internal/govapi.Handler.RegisterRoutes. It does NOT parse
// the YAML or introspect the actual *http.ServeMux (Go's ServeMux exposes no
// API to enumerate registered patterns), so it cannot catch every possible
// drift — but it does fail loudly if a route is added to govapi and this
// list is never updated to match, which is the realistic failure mode.
// Keep this list in sync with internal/govapi/handler.go's RegisterRoutes.
func TestOpenAPISpec_CoversEveryGovAPIRoute(t *testing.T) {
	spec := string(specYAML)

	routes := []string{
		"/api/v1/workspaces/{workspace_id}/policies/validate",
		"/api/v1/workspaces/{workspace_id}/policies/rollback",
		"/api/v1/workspaces/{workspace_id}/policies/{version}/activate",
		"/api/v1/workspaces/{workspace_id}/policies/{version}/preview",
		"/api/v1/workspaces/{workspace_id}/policies/{version}/dryrun",
		"/api/v1/workspaces/{workspace_id}/policies/{version}:",
		"/api/v1/workspaces/{workspace_id}/policies:",
		"/api/v1/workspaces/{workspace_id}/audit-events",
		"/api/v1/workspaces/{workspace_id}/tools",
	}
	for _, route := range routes {
		if !strings.Contains(spec, route) {
			t.Errorf("openapi.yaml is missing a path entry for %q — internal/govapi registers this route; keep the spec in sync", route)
		}
	}
}
