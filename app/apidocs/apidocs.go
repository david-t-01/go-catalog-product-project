package apidocs

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"strings"
)

//go:embed static
var static embed.FS

// Register mounts the OpenAPI spec and the vendored Swagger UI under /docs/.
func Register(mux *http.ServeMux) {
	// Go's built-in mime table has no .yaml entry; without this the spec would be served as text/plain.
	_ = mime.AddExtensionType(".yaml", "application/yaml")
	ui, _ := fs.Sub(static, "static") // constant path inside the embedded tree, cannot fail
	files := http.StripPrefix("/docs/", http.FileServerFS(ui))
	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, r *http.Request) {
		// The 1.7 MB UI bundle changes only with `make docs-ui`; the spec and index must stay fresh.
		if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		files.ServeHTTP(w, r)
	})
}
