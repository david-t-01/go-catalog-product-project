package apidocs

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static
var static embed.FS

//go:embed static/openapi.yaml
var spec []byte

// Register mounts the OpenAPI spec and the vendored Swagger UI under /docs/.
func Register(mux *http.ServeMux) {
	ui, err := fs.Sub(static, "static")
	if err != nil {
		panic("apidocs: static assets missing: " + err.Error())
	}
	mux.HandleFunc("GET /docs/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec)
	})
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServerFS(ui)))
}
