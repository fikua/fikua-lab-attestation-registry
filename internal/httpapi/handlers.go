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
	"github.com/fikua/fikua-lab-attestation-registry/internal/model"
	"github.com/fikua/fikua-lab-attestation-registry/internal/sdjwtvc"
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
// (/api/v1/schemes, /health) are consumed directly by issuer/verifier at
// this service's own hostname and are unaffected by it. Pass "" when
// served at the root (local dev, direct access).
func NewHandler(c *catalogue.Catalogue, spec []byte, basePath string) *Handler {
	return &Handler{catalogue: c, spec: spec, basePath: strings.TrimSuffix(basePath, "/")}
}

// Routes registers this handler's endpoints on mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/schemes", h.listSchemes)
	// {id} matches a single path segment; every scheme id we define uses
	// ":" (URN-style) but never "/", so this is sufficient and — unlike
	// {id...} — allows further static segments (/{version}/type-metadata)
	// after it.
	mux.HandleFunc("GET /api/v1/schemes/{id}", h.getScheme)
	// Versioned, immutable per FR-18/FR-19/FR-20: this is the only route
	// for the Type Metadata Document — there is deliberately no unversioned
	// alias, so a consumer's schemaURIs.uri always names an exact version
	// that will never change contents out from under it. Callers that want
	// "whatever the current version is" read scheme.version from
	// GET /api/v1/schemes/{id} and build this URL themselves.
	mux.HandleFunc("GET /api/v1/schemes/{id}/{version}/type-metadata", h.getTypeMetadata)
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET "+h.basePath+"/openapi.yaml", h.openAPISpec)
	mux.HandleFunc("GET "+h.basePath+"/swagger", h.swaggerUI)
}

func (h *Handler) listSchemes(w http.ResponseWriter, r *http.Request) {
	all := h.catalogue.All()
	out := make([]definitionDTO, len(all))
	for i, d := range all {
		out[i] = h.toDTO(r, d)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) getScheme(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // wildcard match: scheme ids like "urn:eudi:pid:1" contain no reserved chars needing extra decoding
	definition, err := h.catalogue.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, h.toDTO(r, definition))
}

// schemaURI is the TS11 §4.3.2 Schema sub-class: a pointer, not a container.
// {formatIdentifier, uri} only — the claim content lives at uri, in a
// separate, format-native document (§4.3.4), not inlined here. uri carries
// a "#integrity=<sri-value>" suffix (TS11 §4.3.1/§4.3.2: "the URI MAY be
// suffixed with #integrity, the value of which SHALL be an 'integrity
// metadata' string as defined in Section 3 of [W3C SRI]") whenever this
// registry can compute one, so a consumer can verify the fetched document
// without trusting the transport (NFR-S-02). Distinct from SD-JWT VC §7's
// own integrity mechanism, which is a sibling JSON claim inside the token
// itself (`vct#integrity`) — not a URL fragment; that one is the token
// issuer's responsibility (fikua-lab-issuer), not this registry's.
type schemaURI struct {
	FormatIdentifier model.CredentialFormat `json:"formatIdentifier"`
	URI              string                 `json:"uri"`
}

// schemeDTO is AttestationScheme as served over HTTP: SchemaMeta's own
// fields, plus schemaURIs computed from the request's host/basePath (the
// model package does no I/O and doesn't know its own deployed address).
// Schemas (this registry's internal claim content, keyed by format) rides
// along too — TS11 doesn't require it inline, but issuer/verifier consume
// it directly today and losing it would be a breaking API change; schemaURIs
// is what makes the response TS11-shaped without taking that away.
type schemeDTO struct {
	ID                 string                   `json:"id"`
	CatalogueID        string                   `json:"catalogueId,omitempty"`
	Version            string                   `json:"version"`
	RulebookURI        string                   `json:"rulebookUri,omitempty"`
	TrustedAuthorities []model.TrustAuthority   `json:"trustedAuthorities,omitempty"`
	AttestationLoS     model.AssuranceLevel     `json:"attestationLoS"`
	BindingType        model.BindingType        `json:"bindingType"`
	SupportedFormats   []model.CredentialFormat `json:"supportedFormats"`
	SchemaURIs         []schemaURI              `json:"schemaURIs"`
	Schemas            []model.FormatSchema     `json:"schemas"`
}

type definitionDTO struct {
	Rulebook model.AttestationRulebook `json:"rulebook"`
	Scheme   schemeDTO                 `json:"scheme"`
}

func (h *Handler) toDTO(r *http.Request, def model.Definition) definitionDTO {
	uris := make([]schemaURI, 0, len(def.Scheme.Schemas))
	for _, s := range def.Scheme.Schemas {
		uris = append(uris, schemaURI{
			FormatIdentifier: s.Format,
			URI:              h.formatDocumentURI(r, def, s.Format),
		})
	}
	return definitionDTO{
		Rulebook: def.Rulebook,
		Scheme: schemeDTO{
			ID:                 def.Scheme.ID,
			CatalogueID:        def.Scheme.CatalogueID,
			Version:            def.Scheme.Version,
			RulebookURI:        def.Scheme.RulebookURI,
			TrustedAuthorities: def.Scheme.TrustedAuthorities,
			AttestationLoS:     def.Scheme.AttestationLoS,
			BindingType:        def.Scheme.BindingType,
			SupportedFormats:   def.Scheme.SupportedFormats,
			SchemaURIs:         uris,
			Schemas:            def.Scheme.Schemas,
		},
	}
}

// formatDocumentURI is the absolute, version-addressed URL of the
// format-native schema document for one (scheme, format) pair, with a
// "#integrity" suffix when this registry can compute one. SD-JWT VC Type
// Metadata only today; a future ISO 23220-2 DocType endpoint for mso_mdoc
// is not built yet (see docs/compliance/*.md "Open gap: mdoc DocType
// document"), so mdoc formats get an empty uri until that endpoint exists
// rather than a link to nothing.
func (h *Handler) formatDocumentURI(r *http.Request, def model.Definition, format model.CredentialFormat) string {
	if format != model.FormatSDJWTVC {
		return ""
	}
	metadata, err := sdjwtvc.FromScheme(def)
	if err != nil {
		return ""
	}
	integrity, err := sdjwtvc.Integrity(metadata)
	if err != nil {
		return ""
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	base := scheme + "://" + r.Host + h.basePath + "/api/v1/schemes/" + def.Scheme.ID + "/" + def.Scheme.Version + "/type-metadata"
	return base + "#integrity=" + integrity
}

// getTypeMetadata serves the SD-JWT VC Type Metadata Document for a
// scheme's dc+sd-jwt format at the exact version named in the path, per the
// SD-JWT VC spec's "Registry" retrieval method — this registry IS that
// registry. version is validated against the scheme's current version
// rather than tracking prior versions' content: this registry does not yet
// keep old versions addressable (FR-18/19/20 immutability is aspirational
// here, not yet backed by a version store — see docs/compliance/*.md).
// 404s for schemes with no SD-JWT VC format, or for a version that isn't
// the current one.
func (h *Handler) getTypeMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	version := r.PathValue("version")
	definition, err := h.catalogue.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if version != definition.Scheme.Version {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no such version: " + version})
		return
	}
	metadata, err := sdjwtvc.FromScheme(definition)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, metadata)
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
