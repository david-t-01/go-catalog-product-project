package catalog_test

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/david-t-01/go-catalog-product-project/app/catalog"
	"github.com/david-t-01/go-catalog-product-project/app/catalog/mocks"
	"github.com/david-t-01/go-catalog-product-project/models"
)

func newServer(svc catalog.Catalog) *httptest.Server {
	h := catalog.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", h.HandleList)
	mux.HandleFunc("GET /catalog/{code}", h.HandleDetails)
	return httptest.NewServer(mux)
}

func get(t *testing.T, srv *httptest.Server, path string) (int, string) {
	t.Helper()
	resp, err := http.Get(srv.URL + path) //nolint:noctx // test helper
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func TestHandler_HandleList(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		expect     func(m *mocks.MockCatalog)
		wantStatus int
		wantBody   string
	}{
		{"defaults", "/catalog", func(m *mocks.MockCatalog) {
			m.EXPECT().List(mock.Anything, catalog.ListQuery{Offset: 0, Limit: 10}).Return(catalog.ListResult{
				Products: []catalog.Product{{Code: "PROD001", Price: 10.99, Category: catalog.Category{Code: "CLOTHING", Name: "Clothing"}}},
				Total:    1, Offset: 0, Limit: 10,
			}, nil)
		}, 200, `{"products":[{"code":"PROD001","price":10.99,"category":{"code":"CLOTHING","name":"Clothing"}}],"total":1,"offset":0,"limit":10}`},
		{"passes raw params to the service", "/catalog?offset=2&limit=500&category=%20SHOES%20&price_less_than=20", func(m *mocks.MockCatalog) {
			m.EXPECT().List(mock.Anything, catalog.ListQuery{Offset: 2, Limit: 500, CategoryCode: " SHOES ", PriceLessThan: dp("20")}).
				Return(catalog.ListResult{Products: []catalog.Product{}, Total: 0, Offset: 2, Limit: 100}, nil)
		}, 200, `{"products":[],"total":0,"offset":2,"limit":100}`},
		{"empty page is an empty array", "/catalog?offset=1000", func(m *mocks.MockCatalog) {
			m.EXPECT().List(mock.Anything, catalog.ListQuery{Offset: 1000, Limit: 10}).
				Return(catalog.ListResult{Products: []catalog.Product{}, Total: 8, Offset: 1000, Limit: 10}, nil)
		}, 200, `{"products":[],"total":8,"offset":1000,"limit":10}`},
		{"non-numeric limit", "/catalog?limit=abc", nil, 400, `{"error":"invalid limit: must be an integer"}`},
		{"non-numeric offset", "/catalog?offset=1.5", nil, 400, `{"error":"invalid offset: must be an integer"}`},
		{"non-decimal price", "/catalog?price_less_than=ten", nil, 400, `{"error":"invalid price_less_than: must be a decimal number"}`},
		{"service error", "/catalog", func(m *mocks.MockCatalog) {
			m.EXPECT().List(mock.Anything, mock.Anything).Return(catalog.ListResult{}, errors.New("boom"))
		}, 500, `{"error":"internal error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockCatalog(t)
			if tt.expect != nil {
				tt.expect(svc)
			}
			srv := newServer(svc)
			defer srv.Close()

			status, body := get(t, srv, tt.url)

			assert.Equal(t, tt.wantStatus, status)
			assert.JSONEq(t, tt.wantBody, body)
		})
	}
}

func TestHandler_HandleDetails(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		expect     func(m *mocks.MockCatalog)
		wantStatus int
		wantBody   string
	}{
		{"found", "PROD001", func(m *mocks.MockCatalog) {
			m.EXPECT().Details(mock.Anything, "PROD001").Return(catalog.ProductDetails{
				Code: "PROD001", Price: 10.99, Category: catalog.Category{Code: "CLOTHING", Name: "Clothing"},
				Variants: []catalog.Variant{{Name: "Variant A", SKU: "SKU001A", Price: 11.99}, {Name: "Variant B", SKU: "SKU001B", Price: 10.99}},
			}, nil)
		}, 200, `{"code":"PROD001","price":10.99,"category":{"code":"CLOTHING","name":"Clothing"},"variants":[{"name":"Variant A","sku":"SKU001A","price":11.99},{"name":"Variant B","sku":"SKU001B","price":10.99}]}`},
		{"not found", "NOPE", func(m *mocks.MockCatalog) {
			m.EXPECT().Details(mock.Anything, "NOPE").Return(catalog.ProductDetails{}, fmt.Errorf("getting product NOPE: %w", models.ErrNotFound))
		}, 404, `{"error":"product not found"}`},
		{"service error", "PROD001", func(m *mocks.MockCatalog) {
			m.EXPECT().Details(mock.Anything, "PROD001").Return(catalog.ProductDetails{}, errors.New("boom"))
		}, 500, `{"error":"internal error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockCatalog(t)
			tt.expect(svc)
			srv := newServer(svc)
			defer srv.Close()

			status, body := get(t, srv, "/catalog/"+tt.code)

			assert.Equal(t, tt.wantStatus, status)
			assert.JSONEq(t, tt.wantBody, body)
		})
	}
}
