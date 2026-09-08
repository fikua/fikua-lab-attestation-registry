package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/httpapi"
	"github.com/fikua/fikua-lab-attestation-registry/web"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}
	mux := http.NewServeMux()
	httpapi.NewHandler(cat, web.OpenAPISpec, "").Routes(mux)
	return httptest.NewServer(mux)
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestListSchemesReturnsAllTwo(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 {
		t.Fatalf("got %d definitions, want 2", len(body))
	}
}

func TestGetSchemeByIDWithColons(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/urn:eudi:pid:1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestGetUnknownSchemeReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetSchemeIncludesSchemaURIs(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/urn:eudi:pid:1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var body struct {
		Scheme struct {
			SchemaURIs []struct {
				FormatIdentifier string `json:"formatIdentifier"`
				URI              string `json:"uri"`
			} `json:"schemaURIs"`
		} `json:"scheme"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	uris := body.Scheme.SchemaURIs
	if len(uris) != 2 {
		t.Fatalf("got %d schemaURIs, want 2 (dc+sd-jwt, mso_mdoc)", len(uris))
	}

	byFormat := make(map[string]string, len(uris))
	for _, u := range uris {
		byFormat[u.FormatIdentifier] = u.URI
	}

	wantPrefix := srv.URL + "/api/v1/schemes/urn:eudi:pid:1/1.7/type-metadata#integrity=sha256-"
	if got := byFormat["dc+sd-jwt"]; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("dc+sd-jwt uri = %q, want prefix %q", got, wantPrefix)
	}
	// mso_mdoc has no format document generator yet (no ISO 23220-2 DocType
	// endpoint) — uri is empty rather than pointing at a 404.
	if got := byFormat["mso_mdoc"]; got != "" {
		t.Errorf("mso_mdoc uri = %q, want empty (no DocType generator yet)", got)
	}
}

func TestGetTypeMetadataForSdJwtScheme(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/urn:eudi:pid:1/1.7/type-metadata")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["vct"] != "urn:eudi:pid:1" {
		t.Errorf("vct = %v, want urn:eudi:pid:1", body["vct"])
	}
}

func TestGetTypeMetadataForWrongVersionReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/urn:eudi:pid:1/0.1/type-metadata")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestGetTypeMetadataForUnknownSchemeReturns404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/schemes/does-not-exist/1/type-metadata")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpenAPISpecIsServed(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if len(web.OpenAPISpec) == 0 {
		t.Fatal("embedded OpenAPI spec is empty")
	}
}

func TestSwaggerUIPageIsServed(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/swagger")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}
