package categories_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/david-t-01/go-catalog-product-project/app/categories"
	"github.com/david-t-01/go-catalog-product-project/app/categories/mocks"
	"github.com/david-t-01/go-catalog-product-project/models"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		in      categories.CreateInput
		expect  func(m *mocks.MockRepository)
		want    categories.Category
		wantErr string // expected err.Error(); "" means no error
		wantIs  error  // sentinel checked with errors.Is, or nil
	}{
		{"creates trimmed", categories.CreateInput{Code: " BAGS ", Name: " Bags "}, func(m *mocks.MockRepository) {
			m.EXPECT().Create(mock.Anything, &models.Category{Code: "BAGS", Name: "Bags"}).Return(nil)
		}, categories.Category{Code: "BAGS", Name: "Bags"}, "", nil},
		{"missing code", categories.CreateInput{Name: "Bags"}, nil, categories.Category{}, "code is required", nil},
		{"missing name", categories.CreateInput{Code: "BAGS"}, nil, categories.Category{}, "name is required", nil},
		{"code too long", categories.CreateInput{Code: strings.Repeat("A", 33), Name: "Bags"}, nil, categories.Category{}, "code must be at most 32 characters", nil},
		{"name too long", categories.CreateInput{Code: "BAGS", Name: strings.Repeat("a", 257)}, nil, categories.Category{}, "name must be at most 256 characters", nil},
		{"duplicate", categories.CreateInput{Code: "SHOES", Name: "Shoes"}, func(m *mocks.MockRepository) {
			m.EXPECT().Create(mock.Anything, mock.Anything).Return(models.ErrConflict)
		}, categories.Category{}, "creating category: already exists", models.ErrConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRepository(t)
			if tt.expect != nil {
				tt.expect(repo)
			}

			got, err := categories.NewService(repo).Create(context.Background(), tt.in)

			assert.Equal(t, tt.want, got)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
			if tt.wantIs != nil {
				assert.True(t, errors.Is(err, tt.wantIs))
			} else {
				var ve *models.ValidationError
				assert.True(t, errors.As(err, &ve), "validation failures are *models.ValidationError")
			}
		})
	}
}

func TestService_List(t *testing.T) {
	t.Run("maps in order", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().List(mock.Anything).Return([]models.Category{{ID: 1, Code: "CLOTHING", Name: "Clothing"}, {ID: 2, Code: "SHOES", Name: "Shoes"}}, nil)

		got, err := categories.NewService(repo).List(context.Background())

		require.NoError(t, err)
		assert.Equal(t, []categories.Category{{Code: "CLOTHING", Name: "Clothing"}, {Code: "SHOES", Name: "Shoes"}}, got)
	})

	t.Run("empty is an empty slice", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().List(mock.Anything).Return(nil, nil)

		got, err := categories.NewService(repo).List(context.Background())

		require.NoError(t, err)
		assert.Equal(t, []categories.Category{}, got)
	})

	t.Run("repository error is wrapped", func(t *testing.T) {
		repo := mocks.NewMockRepository(t)
		repo.EXPECT().List(mock.Anything).Return(nil, errors.New("boom"))

		_, err := categories.NewService(repo).List(context.Background())

		assert.EqualError(t, err, "listing categories: boom")
	})
}
