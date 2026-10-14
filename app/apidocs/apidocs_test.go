package apidocs_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/david-t-01/go-catalog-product-project/app/apidocs"
)

func get(t *testing.T, srv *httptest.Server, path string) (status int, contentType, body string) {
	t.Helper()
	resp, err := http.Get(srv.URL + path) //nolint:noctx // test helper
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(raw)
}

func TestRegister(t *testing.T) {
	mux := http.NewServeMux()
	apidocs.Register(mux)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	t.Run("spec is served as yaml", func(t *testing.T) {
		status, contentType, body := get(t, srv, "/docs/openapi.yaml")
		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, "application/yaml", contentType)
		assert.Contains(t, body, "/catalog/{code}")
	})

	t.Run("ui is served", func(t *testing.T) {
		status, contentType, body := get(t, srv, "/docs/")
		assert.Equal(t, http.StatusOK, status)
		assert.True(t, strings.HasPrefix(contentType, "text/html"))
		assert.Contains(t, body, `id="swagger-ui"`)
	})

	t.Run("spec covers every API route", func(t *testing.T) {
		_, _, body := get(t, srv, "/docs/openapi.yaml")
		for _, route := range []string{"/healthz:", "/catalog:", "/catalog/{code}:", "/categories:"} {
			assert.Contains(t, body, "  "+route)
		}
	})
}
