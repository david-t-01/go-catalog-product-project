package models

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type ProductsRepository struct {
	db *gorm.DB
}

func NewProductsRepository(db *gorm.DB) *ProductsRepository {
	return &ProductsRepository{db: db}
}

// List returns one page of products matching f, plus the total count before paging.
func (r *ProductsRepository) List(ctx context.Context, f ProductFilter, p Page) ([]Product, int64, error) {
	base := r.db.WithContext(ctx).Model(&Product{})
	if f.CategoryCode != "" {
		base = base.Where("category_id IN (?)", r.db.Model(&Category{}).Select("id").Where("code = ?", f.CategoryCode))
	}
	if f.PriceLessThan != nil {
		base = base.Where("price < ?", *f.PriceLessThan)
	}
	// Session() lets the same filtered query be reused for Count and Find without leaking clauses.
	base = base.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting products: %w", err)
	}

	products := []Product{}
	if err := base.Preload("Category").Order("id").Offset(p.Offset).Limit(p.Limit).Find(&products).Error; err != nil {
		return nil, 0, fmt.Errorf("listing products: %w", err)
	}
	return products, total, nil
}

func (r *ProductsRepository) GetByCode(ctx context.Context, code string) (*Product, error) {
	var p Product
	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Variants", func(db *gorm.DB) *gorm.DB { return db.Order("id") }).
		Where("code = ?", code).
		First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting product %s: %w", code, err)
	}
	return &p, nil
}
