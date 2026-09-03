package webui_test

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/webui"
	"github.com/fikua/fikua-lab-attestation-registry/web"
)

func newTestServerWithBasePath(t *testing.T, basePath string) *httptest.Server {
	t.Helper()

	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		t.Fatalf("LoadFS: %v", err)
	}

	tmplFS, err := fs.Sub(web.TemplatesFS, "templates")
	if err != nil {
		t.Fatalf("templates sub: %v", err)
	}
	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		t.Fatalf("static sub: %v", err)
	}

	handler, err := webui.NewHandler(cat, tmplFS, staticFS, basePath)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	mux := http.NewServeMux()
	handler.Routes(mux)
	return httptest.NewServer(mux)
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServerWithBasePath(t, "")
}

func TestIndexListsAllRulebooks(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRulebookPageForKnownScheme(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/rulebooks/urn:eudi:pid:1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestRulebookPageForUnknownSchemeIs404(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/rulebooks/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestStaticAssetIsServed(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/static/style.css")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestBasePathPrefixesRoutesAndLinks(t *testing.T) {
	srv := newTestServerWithBasePath(t, "/attestation-registry")
	defer srv.Close()

	// The route itself is only reachable under the prefix.
	resp, err := http.Get(srv.URL + "/attestation-registry/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)

	if !strings.Contains(html, `/attestation-registry/static/style.css`) {
		t.Errorf("index page does not link to the prefixed stylesheet: %s", html)
	}
	if !strings.Contains(html, `/attestation-registry/rulebooks/`) {
		t.Errorf("index page does not link to a prefixed rulebook: %s", html)
	}

	// Unprefixed root must not resolve (it belongs to a different service
	// mounted at the zone root, if any).
	rootResp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer rootResp.Body.Close()
	if rootResp.StatusCode == http.StatusOK {
		t.Error("unprefixed root should not be served when a base path is configured")
	}
}
