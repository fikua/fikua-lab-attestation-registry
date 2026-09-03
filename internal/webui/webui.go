// Package webui serves the human-readable view of the Attestation Rulebooks
// catalogue: the Rulebook is meant for people, so it needs a page, not just a
// JSON payload.
package webui

import (
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/fikua/fikua-lab-attestation-registry/internal/catalogue"
)

// Handler serves the server-rendered rulebook browser.
type Handler struct {
	catalogue    *catalogue.Catalogue
	indexTmpl    *template.Template
	rulebookTmpl *template.Template
	staticFS     fs.FS
}

// NewHandler builds a webui Handler. templatesFS must contain layout.html
// plus one file per page ("index.html", "rulebook.html") at its root, each
// defining its own "title" and "content" blocks. staticFS is served under
// /static/. Pages are parsed as separate template sets so their same-named
// "title"/"content" blocks don't collide.
func NewHandler(c *catalogue.Catalogue, templatesFS fs.FS, staticFS fs.FS) (*Handler, error) {
	funcs := template.FuncMap{"join": strings.Join}

	indexTmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "layout.html", "index.html")
	if err != nil {
		return nil, err
	}
	rulebookTmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "layout.html", "rulebook.html")
	if err != nil {
		return nil, err
	}

	return &Handler{catalogue: c, indexTmpl: indexTmpl, rulebookTmpl: rulebookTmpl, staticFS: staticFS}, nil
}

// Routes registers this handler's endpoints on mux, including the embedded
// static assets under /static/.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /rulebooks/{id...}", h.rulebook)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(h.staticFS))))
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	h.render(w, h.indexTmpl, map[string]any{
		"Definitions": h.catalogue.All(),
	})
}

func (h *Handler) rulebook(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	definition, err := h.catalogue.Get(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	h.render(w, h.rulebookTmpl, map[string]any{
		"Definition": definition,
	})
}

func (h *Handler) render(w http.ResponseWriter, tmpl *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
