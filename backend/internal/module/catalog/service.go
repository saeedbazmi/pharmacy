package catalog

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// Service holds the catalog business rules.
type Service struct {
	products ProductReader
	log      *slog.Logger
}

// NewService builds the service from its dependencies.
func NewService(products ProductReader, log *slog.Logger) *Service {
	return &Service{products: products, log: log}
}

// ProductBySlug returns a published product, or ErrProductNotFound.
func (s *Service) ProductBySlug(ctx context.Context, slug string) (Product, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return Product{}, ErrInvalidSlug
	}

	product, err := s.products.ProductBySlug(ctx, slug)
	if err != nil {
		return Product{}, fmt.Errorf("product by slug: %w", err)
	}
	return product, nil
}

// normalizeSlug trims the identifier and rejects obviously unusable input
// before it reaches the database.
func normalizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(slug) == 0 || len(slug) > 200 {
		return ""
	}
	for _, r := range slug {
		switch {
		case r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '-':
		default:
			return ""
		}
	}
	return slug
}
