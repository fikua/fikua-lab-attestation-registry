package httpapi

import (
	"fmt"
	"net/http"
)

// swaggerPageTemplate renders Swagger UI via CDN, pointed at the embedded
// OpenAPI spec served from {basePath}/openapi.yaml. No swagger-ui assets are
// bundled — this keeps the module dependency-free — the CDN script/CSS are
// the only external calls this service ever makes, and only from the
// browser rendering this page, not from the server itself.
const swaggerPageTemplate = `<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>Fikua Attestation Registry &mdash; API docs</title>
  <link rel="icon" type="image/svg+xml" href=%q>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui.min.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-bundle.min.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: %q,
        dom_id: '#swagger-ui',
        presets: [SwaggerUIBundle.presets.apis],
      });
    };
  </script>
</body>
</html>`

func (h *Handler) swaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := fmt.Sprintf(swaggerPageTemplate, h.basePath+"/static/favicon.svg", h.basePath+"/openapi.yaml")
	_, _ = w.Write([]byte(page))
}

func (h *Handler) openAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(h.spec)
}
