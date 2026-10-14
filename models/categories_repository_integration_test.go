//go:build integration

package models

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCategoriesRepository(t *testing.T) {
	db := testDB(t)
	repo := NewCategoriesRepository(db)
	ctx := context.Background()

	t.Run("list seeded categories in id order", func(t *testing.T) {
		got, err := repo.List(ctx)
		require.NoError(t, err)
		codes := make([]string, 0, len(got))
		for _, c := range got {
			codes = append(codes, c.Code)
		}
		assert.Equal(t, []string{"CLOTHING", "SHOES", "ACCESSORIES"}, codes)
	})

	t.Run("create assigns an id", func(t *testing.T) {
		c := &Category{Code: "BAGS", Name: "Bags"}
		t.Cleanup(func() { db.Where("code = ?", "BAGS").Delete(&Category{}) })
		require.NoError(t, repo.Create(ctx, c))
		assert.NotZero(t, c.ID)
	})

	t.Run("duplicate code is a conflict", func(t *testing.T) {
		err := repo.Create(ctx, &Category{Code: "SHOES", Name: "Dup"})
		assert.True(t, errors.Is(err, ErrConflict))
	})
}
