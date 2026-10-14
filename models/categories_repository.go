package models

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type CategoriesRepository struct {
	db *gorm.DB
}

func NewCategoriesRepository(db *gorm.DB) *CategoriesRepository {
	return &CategoriesRepository{db: db}
}

func (r *CategoriesRepository) List(ctx context.Context) ([]Category, error) {
	categories := []Category{}
	if err := r.db.WithContext(ctx).Order("id").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	return categories, nil
}

func (r *CategoriesRepository) Create(ctx context.Context, c *Category) error {
	err := r.db.WithContext(ctx).Create(c).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return ErrConflict
	}
	if err != nil {
		return fmt.Errorf("creating category: %w", err)
	}
	return nil
}
