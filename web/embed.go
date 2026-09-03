// Package web embeds the human-facing UI's templates and static assets so
// they ship inside the compiled binary.
package web

import "embed"

//go:embed templates/*.html
var TemplatesFS embed.FS

//go:embed static/*
var StaticFS embed.FS
