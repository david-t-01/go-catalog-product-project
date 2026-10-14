package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/david-t-01/go-catalog-product-project/models"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

// ListQuery carries the raw list parameters; the service applies clamping and trimming.
type ListQuery struct {
	Offset        int
	Limit         int
	CategoryCode  string
	PriceLessThan *decimal.Decimal
}

type ProductsRepository interface {
	List(ctx context.Context, f models.ProductFilter, p models.Page) ([]models.Product, int64, error)
	GetByCode(ctx context.Context, code string) (*models.Product, error)
}

type Service struct {
	repo ProductsRepository
}

func NewService(repo ProductsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
	page := models.Page{
		Offset: max(q.Offset, 0),
		Limit:  min(max(q.Limit, 1), MaxLimit),
	}
	filter := models.ProductFilter{
		CategoryCode:  strings.TrimSpace(q.CategoryCode),
		PriceLessThan: q.PriceLessThan,
	}

	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return ListResult{}, fmt.Errorf("listing products: %w", err)
	}

	products := make([]Product, 0, len(items))
	for _, p := range items {
		products = append(products, toProduct(p))
	}
	return ListResult{Products: products, Total: total, Offset: page.Offset, Limit: page.Limit}, nil
}

func toCategory(c models.Category) Category {
	return Category{Code: c.Code, Name: c.Name}
}

func toProduct(p models.Product) Product {
	return Product{Code: p.Code, Price: p.Price.InexactFloat64(), Category: toCategory(p.Category)}
}

func (s *Service) Details(ctx context.Context, code string) (ProductDetails, error) {
	p, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return ProductDetails{}, fmt.Errorf("getting product %s: %w", code, err)
	}
	return toDetails(p), nil
}

func toDetails(p *models.Product) ProductDetails {
	variants := make([]Variant, 0, len(p.Variants))
	for _, v := range p.Variants {
		variants = append(variants, Variant{Name: v.Name, SKU: v.SKU, Price: v.EffectivePrice(p.Price).InexactFloat64()})
	}
	return ProductDetails{Code: p.Code, Price: p.Price.InexactFloat64(), Category: toCategory(p.Category), Variants: variants}
}
