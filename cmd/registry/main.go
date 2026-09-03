// Command registry runs the Fikua Attestation Registry: a self-contained
// service exposing the ARF 3.0 / ETSI TS 119 472 Catalogue of Attestations
// over a JSON API (for fikua-issuer and fikua-verifier) and a human-readable
// rulebook browser.
package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/fikua/fikua-lab-attestation-registry/data/attestations"
	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
	"github.com/fikua/fikua-lab-attestation-registry/internal/httpapi"
	"github.com/fikua/fikua-lab-attestation-registry/internal/webui"
	"github.com/fikua/fikua-lab-attestation-registry/web"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	// Set when this service is reached through a reverse-proxying Worker
	// under a path prefix (e.g. "/attestation-registry" for
	// lab.fikua.com/attestation-registry/) so its own HTML/redirects point
	// back through that prefix. Empty for local dev / direct access.
	basePath := os.Getenv("BASE_PATH")

	cat, err := catalogue.LoadFS(attestations.FS, ".")
	if err != nil {
		log.Fatalf("loading attestation catalogue: %v", err)
	}

	tmplFS, err := fs.Sub(web.TemplatesFS, "templates")
	if err != nil {
		log.Fatalf("templates: %v", err)
	}
	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		log.Fatalf("static assets: %v", err)
	}
	ui, err := webui.NewHandler(cat, tmplFS, staticFS, basePath)
	if err != nil {
		log.Fatalf("building web UI: %v", err)
	}

	mux := http.NewServeMux()
	httpapi.NewHandler(cat, web.OpenAPISpec, basePath).Routes(mux)
	ui.Routes(mux)

	log.Printf("fikua-lab-attestation-registry listening on %s (%d attestation definitions loaded)", addr, len(cat.All()))
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
