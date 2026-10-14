package categories_test

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/david-t-01/go-catalog-product-project/app/categories"
	"github.com/david-t-01/go-catalog-product-project/app/categories/mocks"
	"github.com/david-t-01/go-catalog-product-project/models"
)

func newServer(svc categories.Categories) *httptest.Server {
	h := categories.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /categories", h.HandleList)
	mux.HandleFunc("POST /categories", h.HandleCreate)
	return httptest.NewServer(mux)
}

func do(t *testing.T, req *http.Request) (int, string) {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, string(body)
}

func TestHandler_HandleList(t *testing.T) {
	tests := []struct {
		name       string
		expect     func(m *mocks.MockCategories)
		wantStatus int
		wantBody   string
	}{
		{"lists", func(m *mocks.MockCategories) {
			m.EXPECT().List(mock.Anything).Return([]categories.Category{{Code: "CLOTHING", Name: "Clothing"}, {Code: "SHOES", Name: "Shoes"}}, nil)
		}, 200, `{"categories":[{"code":"CLOTHING","name":"Clothing"},{"code":"SHOES","name":"Shoes"}]}`},
		{"empty", func(m *mocks.MockCategories) {
			m.EXPECT().List(mock.Anything).Return([]categories.Category{}, nil)
		}, 200, `{"categories":[]}`},
		{"service error", func(m *mocks.MockCategories) {
			m.EXPECT().List(mock.Anything).Return(nil, errors.New("boom"))
		}, 500, `{"error":"internal error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockCategories(t)
			tt.expect(svc)
			srv := newServer(svc)
			defer srv.Close()
			req, err := http.NewRequest(http.MethodGet, srv.URL+"/categories", nil) //nolint:noctx // test helper
			require.NoError(t, err)

			status, body := do(t, req)

			assert.Equal(t, tt.wantStatus, status)
			assert.JSONEq(t, tt.wantBody, body)
		})
	}
}

func TestHandler_HandleCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		expect     func(m *mocks.MockCategories)
		wantStatus int
		wantBody   string
	}{
		{"created", `{"code":"BAGS","name":"Bags"}`, func(m *mocks.MockCategories) {
			m.EXPECT().Create(mock.Anything, categories.CreateInput{Code: "BAGS", Name: "Bags"}).Return(categories.Category{Code: "BAGS", Name: "Bags"}, nil)
		}, 201, `{"code":"BAGS","name":"Bags"}`},
		{"id in body is ignored", `{"id":99,"code":"BAGS","name":"Bags"}`, func(m *mocks.MockCategories) {
			m.EXPECT().Create(mock.Anything, categories.CreateInput{Code: "BAGS", Name: "Bags"}).Return(categories.Category{Code: "BAGS", Name: "Bags"}, nil)
		}, 201, `{"code":"BAGS","name":"Bags"}`},
		{"validation error", `{"name":"Bags"}`, func(m *mocks.MockCategories) {
			m.EXPECT().Create(mock.Anything, categories.CreateInput{Name: "Bags"}).Return(categories.Category{}, &models.ValidationError{Field: "code", Msg: "is required"})
		}, 400, `{"error":"code is required"}`},
		{"malformed json", `{"code":`, nil, 400, `{"error":"invalid JSON body"}`},
		{"duplicate", `{"code":"SHOES","name":"Shoes"}`, func(m *mocks.MockCategories) {
			m.EXPECT().Create(mock.Anything, categories.CreateInput{Code: "SHOES", Name: "Shoes"}).Return(categories.Category{}, fmt.Errorf("creating category: %w", models.ErrConflict))
		}, 409, `{"error":"category code already exists"}`},
		{"service error", `{"code":"BAGS","name":"Bags"}`, func(m *mocks.MockCategories) {
			m.EXPECT().Create(mock.Anything, mock.Anything).Return(categories.Category{}, errors.New("boom"))
		}, 500, `{"error":"internal error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := mocks.NewMockCategories(t)
			if tt.expect != nil {
				tt.expect(svc)
			}
			srv := newServer(svc)
			defer srv.Close()
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/categories", strings.NewReader(tt.body)) //nolint:noctx // test helper
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			status, body := do(t, req)

			assert.Equal(t, tt.wantStatus, status)
			assert.JSONEq(t, tt.wantBody, body)
		})
	}
}
