// Package httpapi exposes the Catalogue of Attestations as a JSON API for
// machine consumers (fikua-issuer, fikua-verifier), following the read
// surface of ARF 3.0 / TS11 §5 (GET-only; this registry does not implement
// the write/PUT/DELETE management API).
package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
)

// Handler serves the attestation catalogue over HTTP.
type Handler struct {
	catalogue *catalogue.Catalogue
	spec      []byte
	basePath  string
}

// NewHandler builds an httpapi Handler. spec is the embedded OpenAPI
// document served at {basePath}/openapi.yaml and rendered by the
// {basePath}/swagger page. basePath is only relevant to the browser-facing
// /swagger page (which needs to know where to fetch the spec from when
// reached through a reverse-proxying Worker); the JSON endpoints
// (/api/v1/schemes, /healthz) are consumed directly by issuer/verifier at
// this service's own hostname and are unaffected by it. Pass "" when
// served at the root (local dev, direct access).
func NewHandler(c *catalogue.Catalogue, spec []byte, basePath string) *Handler {
	return &Handler{catalogue: c, spec: spec, basePath: strings.TrimSuffix(basePath, "/")}
}

// Routes registers this handler's endpoints on mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/schemes", h.listSchemes)
	mux.HandleFunc("GET /api/v1/schemes/{id...}", h.getScheme)
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET "+h.basePath+"/openapi.yaml", h.openAPISpec)
	mux.HandleFunc("GET "+h.basePath+"/swagger", h.swaggerUI)
}

func (h *Handler) listSchemes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.catalogue.All())
}

func (h *Handler) getScheme(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // wildcard match: scheme ids like "urn:eudi:pid:1" contain no reserved chars needing extra decoding
	definition, err := h.catalogue.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, definition)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
