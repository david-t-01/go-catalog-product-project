//go:build integration

package models

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductsRepository_List(t *testing.T) {
	repo := NewProductsRepository(testDB(t))
	tests := []struct {
		name      string
		filter    ProductFilter
		page      Page
		wantCodes []string
		wantTotal int64
	}{
		{"first page", ProductFilter{}, Page{0, 10}, []string{"PROD001", "PROD002", "PROD003", "PROD004", "PROD005", "PROD006", "PROD007", "PROD008"}, 8},
		{"second page of 3", ProductFilter{}, Page{3, 3}, []string{"PROD004", "PROD005", "PROD006"}, 8},
		{"shoes", ProductFilter{CategoryCode: "SHOES"}, Page{0, 10}, []string{"PROD002", "PROD006"}, 2},
		{"price below 10", ProductFilter{PriceLessThan: dec("10")}, Page{0, 10}, []string{"PROD003", "PROD006", "PROD008"}, 3},
		{"price below 0", ProductFilter{PriceLessThan: dec("0")}, Page{0, 10}, []string{}, 0},
		{"unknown category", ProductFilter{CategoryCode: "NOPE"}, Page{0, 10}, []string{}, 0},
		{"category and price", ProductFilter{CategoryCode: "CLOTHING", PriceLessThan: dec("15")}, Page{0, 10}, []string{"PROD001"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			products, total, err := repo.List(context.Background(), tt.filter, tt.page)
			require.NoError(t, err)
			require.NotNil(t, products)
			codes := make([]string, 0, len(products))
			for _, p := range products {
				codes = append(codes, p.Code)
				assert.NotEmpty(t, p.Category.Code, "category must be preloaded")
			}
			assert.Equal(t, tt.wantCodes, codes)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestProductsRepository_GetByCode(t *testing.T) {
	repo := NewProductsRepository(testDB(t))

	t.Run("found with variants and category", func(t *testing.T) {
		p, err := repo.GetByCode(context.Background(), "PROD001")
		require.NoError(t, err)
		assert.Equal(t, "CLOTHING", p.Category.Code)
		require.Len(t, p.Variants, 3)
		assert.False(t, p.Variants[1].Price.Valid, "Variant B has no own price")
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByCode(context.Background(), "NOPE")
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}
