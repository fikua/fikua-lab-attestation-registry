// Package httpapi exposes the Catalogue of Attestations as a JSON API for
// machine consumers (fikua-issuer, fikua-verifier), following the read
// surface of ARF 3.0 / TS11 §5 (GET-only; this registry does not implement
// the write/PUT/DELETE management API).
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
)

// Handler serves the attestation catalogue over HTTP.
type Handler struct {
	catalogue *catalogue.Catalogue
}

func NewHandler(c *catalogue.Catalogue) *Handler {
	return &Handler{catalogue: c}
}

// Routes registers this handler's endpoints on mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/schemes", h.listSchemes)
	mux.HandleFunc("GET /api/v1/schemes/{id...}", h.getScheme)
	mux.HandleFunc("GET /healthz", h.health)
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
