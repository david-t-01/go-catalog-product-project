package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/david-t-01/go-catalog-product-project/app/catalog"
	"github.com/david-t-01/go-catalog-product-project/app/catalog/mocks"
	"github.com/david-t-01/go-catalog-product-project/models"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func dp(s string) *decimal.Decimal {
	v := d(s)
	return &v
}

func TestCatalogService_List(t *testing.T) {
	clothing := models.Category{Code: "CLOTHING", Name: "Clothing"}
	tests := []struct {
		name     string
		query    catalog.ListQuery
		wantFilt models.ProductFilter
		wantPage models.Page
		repoRet  []models.Product
		repoTot  int64
		want     catalog.ListResult
	}{
		{"defaults pass through", catalog.ListQuery{Offset: 0, Limit: catalog.DefaultLimit},
			models.ProductFilter{}, models.Page{Offset: 0, Limit: 10},
			[]models.Product{{Code: "PROD001", Price: d("10.99"), Category: clothing}}, 1,
			catalog.ListResult{Products: []catalog.Product{{Code: "PROD001", Price: 10.99, Category: catalog.Category{Code: "CLOTHING", Name: "Clothing"}}}, Total: 1, Offset: 0, Limit: 10}},
		{"clamps high limit and negative offset", catalog.ListQuery{Offset: -3, Limit: 500},
			models.ProductFilter{}, models.Page{Offset: 0, Limit: 100}, nil, 8,
			catalog.ListResult{Products: []catalog.Product{}, Total: 8, Offset: 0, Limit: 100}},
		{"clamps zero limit", catalog.ListQuery{Limit: 0},
			models.ProductFilter{}, models.Page{Offset: 0, Limit: 1}, nil, 8,
			catalog.ListResult{Products: []catalog.Product{}, Total: 8, Offset: 0, Limit: 1}},
		{"trims category and passes price", catalog.ListQuery{Limit: 10, CategoryCode: "  SHOES ", PriceLessThan: dp("20")},
			models.ProductFilter{CategoryCode: "SHOES", PriceLessThan: dp("20")}, models.Page{Offset: 0, Limit: 10}, nil, 0,
			catalog.ListResult{Products: []catalog.Product{}, Total: 0, Offset: 0, Limit: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockProductsRepository(t)
			repo.EXPECT().List(mock.Anything, tt.wantFilt, tt.wantPage).Return(tt.repoRet, tt.repoTot, nil)

			got, err := catalog.NewService(repo).List(context.Background(), tt.query)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.NotNil(t, got.Products)
		})
	}

	t.Run("repository error is wrapped", func(t *testing.T) {
		repo := mocks.NewMockProductsRepository(t)
		repo.EXPECT().List(mock.Anything, mock.Anything, mock.Anything).Return(nil, 0, errors.New("boom"))

		_, err := catalog.NewService(repo).List(context.Background(), catalog.ListQuery{Limit: 10})

		assert.ErrorContains(t, err, "listing products: boom")
	})
}

func TestCatalogService_Details(t *testing.T) {
	clothing := models.Category{Code: "CLOTHING", Name: "Clothing"}

	t.Run("maps variants with inherited price", func(t *testing.T) {
		repo := mocks.NewMockProductsRepository(t)
		repo.EXPECT().GetByCode(mock.Anything, "PROD001").Return(&models.Product{
			Code: "PROD001", Price: d("10.99"), Category: clothing,
			Variants: []models.Variant{
				{Name: "Variant A", SKU: "SKU001A", Price: decimal.NewNullDecimal(d("11.99"))},
				{Name: "Variant B", SKU: "SKU001B"},
			},
		}, nil)

		got, err := catalog.NewService(repo).Details(context.Background(), "PROD001")

		require.NoError(t, err)
		assert.Equal(t, catalog.ProductDetails{
			Code: "PROD001", Price: 10.99, Category: catalog.Category{Code: "CLOTHING", Name: "Clothing"},
			Variants: []catalog.Variant{{Name: "Variant A", SKU: "SKU001A", Price: 11.99}, {Name: "Variant B", SKU: "SKU001B", Price: 10.99}},
		}, got)
	})

	t.Run("no variants is an empty slice", func(t *testing.T) {
		repo := mocks.NewMockProductsRepository(t)
		repo.EXPECT().GetByCode(mock.Anything, "PROD006").Return(&models.Product{Code: "PROD006", Price: d("5.50"), Category: clothing}, nil)

		got, err := catalog.NewService(repo).Details(context.Background(), "PROD006")

		require.NoError(t, err)
		assert.Equal(t, []catalog.Variant{}, got.Variants)
	})

	t.Run("not found passes through", func(t *testing.T) {
		repo := mocks.NewMockProductsRepository(t)
		repo.EXPECT().GetByCode(mock.Anything, "NOPE").Return(nil, models.ErrNotFound)

		_, err := catalog.NewService(repo).Details(context.Background(), "NOPE")

		assert.True(t, errors.Is(err, models.ErrNotFound))
	})
}
