package categories

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/david-t-01/go-catalog-product-project/models"
)

const (
	maxCodeLen = 32
	maxNameLen = 256
)

type Repository interface {
	List(ctx context.Context) ([]models.Category, error)
	Create(ctx context.Context, c *models.Category) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Category, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}
	out := make([]Category, 0, len(items))
	for _, c := range items {
		out = append(out, Category{Code: c.Code, Name: c.Name})
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, in CreateInput) (Category, error) {
	code := strings.TrimSpace(in.Code)
	name := strings.TrimSpace(in.Name)
	switch {
	case code == "":
		return Category{}, &models.ValidationError{Field: "code", Msg: "is required"}
	case name == "":
		return Category{}, &models.ValidationError{Field: "name", Msg: "is required"}
	case utf8.RuneCountInString(code) > maxCodeLen:
		return Category{}, &models.ValidationError{Field: "code", Msg: fmt.Sprintf("must be at most %d characters", maxCodeLen)}
	case utf8.RuneCountInString(name) > maxNameLen:
		return Category{}, &models.ValidationError{Field: "name", Msg: fmt.Sprintf("must be at most %d characters", maxNameLen)}
	}

	// The ID is always assigned by the database; input never carries one.
	c := &models.Category{Code: code, Name: name}
	if err := s.repo.Create(ctx, c); err != nil {
		return Category{}, fmt.Errorf("creating category: %w", err)
	}
	return Category{Code: c.Code, Name: c.Name}, nil
}
