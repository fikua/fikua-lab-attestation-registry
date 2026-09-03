package webui_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/webui"
	"github.com/fikua/fikua-lab-attestation-registry/web"
)

func newTestServer(t *testing.T) *httptest.Server {
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

	handler, err := webui.NewHandler(cat, tmplFS, staticFS)
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}

	mux := http.NewServeMux()
	handler.Routes(mux)
	return httptest.NewServer(mux)
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
