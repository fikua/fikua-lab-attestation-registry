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
	basePath     string
}

// NewHandler builds a webui Handler. templatesFS must contain layout.html
// plus one file per page ("index.html", "rulebook.html") at its root, each
// defining its own "title" and "content" blocks. staticFS is served under
// {basePath}/static/. Pages are parsed as separate template sets so their
// same-named "title"/"content" blocks don't collide.
//
// basePath is prepended to every link this handler renders or redirects to
// (e.g. "/static/style.css" becomes "/attestation-registry/static/style.css").
// It's needed because this service is reachable through a reverse-proxying
// Worker at lab.fikua.com/attestation-registry/ as well as directly at its
// own root — pass "" when served at the root (local dev, direct access).
func NewHandler(c *catalogue.Catalogue, templatesFS fs.FS, staticFS fs.FS, basePath string) (*Handler, error) {
	basePath = strings.TrimSuffix(basePath, "/")
	funcs := template.FuncMap{
		"join": strings.Join,
		"base": func(path string) string { return basePath + path },
	}

	indexTmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "layout.html", "index.html")
	if err != nil {
		return nil, err
	}
	rulebookTmpl, err := template.New("").Funcs(funcs).ParseFS(templatesFS, "layout.html", "rulebook.html")
	if err != nil {
		return nil, err
	}

	return &Handler{catalogue: c, indexTmpl: indexTmpl, rulebookTmpl: rulebookTmpl, staticFS: staticFS, basePath: basePath}, nil
}

// Routes registers this handler's endpoints on mux, including the embedded
// static assets under {basePath}/static/.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+h.basePath+"/{$}", h.index)
	mux.HandleFunc("GET "+h.basePath+"/rulebooks/{id...}", h.rulebook)
	mux.Handle("GET "+h.basePath+"/static/", http.StripPrefix(h.basePath+"/static/", http.FileServer(http.FS(h.staticFS))))
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	h.render(w, h.indexTmpl, map[string]any{
		"Definitions": h.catalogue.All(),
		"BasePath":    h.basePath,
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
		"BasePath":   h.basePath,
	})
}

func (h *Handler) render(w http.ResponseWriter, tmpl *template.Template, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
