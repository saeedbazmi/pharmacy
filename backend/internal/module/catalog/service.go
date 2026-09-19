package catalog

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"
)

// Service holds the catalog business rules.
type Service struct {
	products      ProductReader
	log           *slog.Logger
	now           func() time.Time
	staleAfter    time.Duration
	criticalAfter time.Duration
}

// NewService builds the service from its dependencies.
func NewService(products ProductReader, log *slog.Logger) *Service {
	return &Service{
		products:      products,
		log:           log,
		now:           time.Now,
		staleAfter:    DefaultStaleAfter,
		criticalAfter: DefaultCriticalAfter,
	}
}

// SetFreshness overrides stale and critical offer ages from process config.
func (s *Service) SetFreshness(staleAfter, criticalAfter time.Duration) {
	if staleAfter > 0 {
		s.staleAfter = staleAfter
	}
	if criticalAfter > 0 {
		s.criticalAfter = criticalAfter
	}
}

// ProductBySlug returns a published product, or ErrProductNotFound.
func (s *Service) ProductBySlug(ctx context.Context, slug string, filter OfferFilter) (Product, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return Product{}, ErrInvalidSlug
	}

	product, err := s.products.ProductBySlug(ctx, slug)
	if err != nil {
		return Product{}, fmt.Errorf("product by slug: %w", err)
	}
	if filter.Brand != "" && !strings.EqualFold(strings.TrimSpace(product.BrandName), strings.TrimSpace(filter.Brand)) {
		product.Offers = nil
		return product, nil
	}
	product.Offers = applyOfferFilter(product.Offers, filter, s.now(), s.staleAfter, s.criticalAfter)
	product.Offers = rankByPrice(product.Offers)
	if filter.Sort == SortPriceDesc {
		reverseOffers(product.Offers)
	}
	return product, nil
}

// ListPublished returns compact catalogue rows for the home and compare flows.
func (s *Service) ListPublished(ctx context.Context, limit int) ([]ProductSummary, error) {
	if limit < 1 || limit > 48 {
		limit = 48
	}
	items, err := s.products.ListPublished(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("list published products: %w", err)
	}
	return items, nil
}

// ListSitemap returns published slugs for the public sitemap.
func (s *Service) ListSitemap(ctx context.Context) ([]SitemapEntry, error) {
	return s.products.ListSitemap(ctx)
}

// ListCategorySitemap returns category slugs that have at least one published product.
func (s *Service) ListCategorySitemap(ctx context.Context) ([]SitemapEntry, error) {
	return s.products.ListCategorySitemap(ctx)
}

// CategoryBySlug returns a live category or ErrCategoryNotFound.
func (s *Service) CategoryBySlug(ctx context.Context, slug string) (Category, error) {
	slug = normalizeSlug(slug)
	if slug == "" {
		return Category{}, ErrInvalidSlug
	}
	got, err := s.products.CategoryBySlug(ctx, slug)
	if err != nil {
		return Category{}, fmt.Errorf("category by slug: %w", err)
	}
	return got, nil
}

// ListByCategory returns published products in a category.
func (s *Service) ListByCategory(ctx context.Context, slug string, limit int) (Category, []ProductSummary, error) {
	cat, err := s.CategoryBySlug(ctx, slug)
	if err != nil {
		return Category{}, nil, err
	}
	if limit < 1 || limit > 48 {
		limit = 48
	}
	items, err := s.products.ListByCategory(ctx, cat.Slug, int32(limit))
	if err != nil {
		return Category{}, nil, fmt.Errorf("list by category: %w", err)
	}
	return cat, items, nil
}

// ProductsBySlugs returns summaries for the compare table, ignoring unknown slugs.
func (s *Service) ProductsBySlugs(ctx context.Context, slugs []string) ([]ProductSummary, error) {
	clean := make([]string, 0, len(slugs))
	seen := map[string]bool{}
	for _, slug := range slugs {
		slug = normalizeSlug(slug)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		clean = append(clean, slug)
	}
	if len(clean) == 0 {
		return []ProductSummary{}, nil
	}
	items, err := s.products.ProductsBySlugs(ctx, clean)
	if err != nil {
		return nil, fmt.Errorf("products by slugs: %w", err)
	}
	return items, nil
}

func rankByPrice(offers []Offer) []Offer {
	if len(offers) == 0 {
		return offers
	}
	sort.SliceStable(offers, func(i, j int) bool {
		if offers[i].PriceRial != offers[j].PriceRial {
			return offers[i].PriceRial < offers[j].PriceRial
		}
		if !offers[i].LastSeenAt.Equal(offers[j].LastSeenAt) {
			return offers[i].LastSeenAt.After(offers[j].LastSeenAt)
		}
		return offers[i].PharmacyName < offers[j].PharmacyName
	})
	for i := range offers {
		offers[i].BestPrice = false
	}
	offers[0].BestPrice = true
	return offers
}

func reverseOffers(offers []Offer) {
	for i, j := 0, len(offers)-1; i < j; i, j = i+1, j-1 {
		offers[i], offers[j] = offers[j], offers[i]
	}
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
