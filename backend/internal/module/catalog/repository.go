package catalog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

// ProductReader is the read side the service depends on. Declared here, on the
// consumer side, so the service never learns about pgx or generated code.
type ProductReader interface {
	ProductBySlug(ctx context.Context, slug string) (Product, error)
	ListPublished(ctx context.Context, limit int32) ([]ProductSummary, error)
	ProductsBySlugs(ctx context.Context, slugs []string) ([]ProductSummary, error)
	ListSitemap(ctx context.Context) ([]SitemapEntry, error)
	ListCategorySitemap(ctx context.Context) ([]SitemapEntry, error)
	CategoryBySlug(ctx context.Context, slug string) (Category, error)
	ListByCategory(ctx context.Context, slug string, limit int32) ([]ProductSummary, error)
}

// Repository reads catalog rows through the generated queries.
type Repository struct {
	q          *store.Queries
	freshSince func() time.Time
}

// NewRepository wires the repository to a pool, connection or transaction.
func NewRepository(db store.DBTX) *Repository {
	return NewRepositoryWithFreshness(db, DefaultCriticalAfter)
}

// NewRepositoryWithFreshness excludes critically stale offers from list aggregates.
func NewRepositoryWithFreshness(db store.DBTX, criticalAfter time.Duration) *Repository {
	if criticalAfter <= 0 {
		criticalAfter = DefaultCriticalAfter
	}
	return &Repository{
		q: store.New(db),
		freshSince: func() time.Time {
			return time.Now().Add(-criticalAfter)
		},
	}
}

// ProductBySlug returns the published product with the given slug.
func (r *Repository) ProductBySlug(ctx context.Context, slug string) (Product, error) {
	row, err := r.q.GetProductBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, fmt.Errorf("query product by slug %q: %w", slug, err)
	}

	product := Product{
		ID:           row.ID,
		Slug:         row.Slug,
		NameFa:       row.NameFa,
		NameEn:       text(row.NameEn),
		GenericName:  text(row.GenericName),
		DosageForm:   text(row.DosageForm),
		Strength:     text(row.Strength),
		ImageURL:     text(row.ImageUrl),
		Description:  text(row.Description),
		BrandName:    text(row.BrandName),
		CategoryName: text(row.CategoryName),
		UpdatedAt:    row.UpdatedAt,
	}

	offerRows, err := r.q.ListActiveOffersByProduct(ctx, row.ID)
	if err != nil {
		return Product{}, fmt.Errorf("list offers for product %d: %w", row.ID, err)
	}
	product.Offers = make([]Offer, 0, len(offerRows))
	for _, o := range offerRows {
		product.Offers = append(product.Offers, Offer{
			ID:           o.ID,
			PriceRial:    o.PriceRial,
			InStock:      o.InStock,
			ProductURL:   o.ProductUrl,
			LastSeenAt:   o.LastSeenAt,
			PharmacyName: o.PharmacyName,
			PharmacySlug: o.PharmacySlug,
		})
	}
	return product, nil
}

func (r *Repository) ListPublished(ctx context.Context, limit int32) ([]ProductSummary, error) {
	rows, err := r.q.ListPublishedProducts(ctx, store.ListPublishedProductsParams{
		FreshSince: r.freshSince(),
		PageLimit:  limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list published products: %w", err)
	}
	out := make([]ProductSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProductSummary{
			ID:              row.ID,
			Slug:            row.Slug,
			NameFa:          row.NameFa,
			ImageURL:        text(row.ImageUrl),
			BrandName:       text(row.BrandName),
			LowestPriceRial: row.LowestPriceRial,
			OfferCount:      int(row.OfferCount),
			InStockCount:    int(row.InStockCount),
		})
	}
	return out, nil
}

func (r *Repository) ProductsBySlugs(ctx context.Context, slugs []string) ([]ProductSummary, error) {
	rows, err := r.q.GetPublishedProductsBySlugs(ctx, store.GetPublishedProductsBySlugsParams{
		Slugs:      slugs,
		FreshSince: r.freshSince(),
	})
	if err != nil {
		return nil, fmt.Errorf("products by slugs: %w", err)
	}
	out := make([]ProductSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProductSummary{
			ID:              row.ID,
			Slug:            row.Slug,
			NameFa:          row.NameFa,
			ImageURL:        text(row.ImageUrl),
			BrandName:       text(row.BrandName),
			LowestPriceRial: row.LowestPriceRial,
			OfferCount:      int(row.OfferCount),
			InStockCount:    int(row.InStockCount),
		})
	}
	return out, nil
}

func (r *Repository) ListSitemap(ctx context.Context) ([]SitemapEntry, error) {
	rows, err := r.q.ListPublishedSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sitemap slugs: %w", err)
	}
	out := make([]SitemapEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, SitemapEntry{Slug: row.Slug, UpdatedAt: row.UpdatedAt})
	}
	return out, nil
}

func (r *Repository) ListCategorySitemap(ctx context.Context) ([]SitemapEntry, error) {
	rows, err := r.q.ListPublishedCategorySlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list category sitemap: %w", err)
	}
	out := make([]SitemapEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, SitemapEntry{Slug: row.Slug, UpdatedAt: row.UpdatedAt})
	}
	return out, nil
}

func (r *Repository) CategoryBySlug(ctx context.Context, slug string) (Category, error) {
	row, err := r.q.GetPublishedCategoryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Category{}, ErrCategoryNotFound
		}
		return Category{}, fmt.Errorf("category by slug %q: %w", slug, err)
	}
	return Category{ID: row.ID, Slug: row.Slug, NameFa: row.NameFa, UpdatedAt: row.UpdatedAt}, nil
}

func (r *Repository) ListByCategory(ctx context.Context, slug string, limit int32) ([]ProductSummary, error) {
	rows, err := r.q.ListPublishedProductsByCategorySlug(ctx, store.ListPublishedProductsByCategorySlugParams{
		FreshSince:   r.freshSince(),
		CategorySlug: slug,
		PageLimit:    limit,
	})
	if err != nil {
		return nil, fmt.Errorf("list products in category %q: %w", slug, err)
	}
	out := make([]ProductSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, ProductSummary{
			ID:              row.ID,
			Slug:            row.Slug,
			NameFa:          row.NameFa,
			ImageURL:        text(row.ImageUrl),
			BrandName:       text(row.BrandName),
			LowestPriceRial: row.LowestPriceRial,
			OfferCount:      int(row.OfferCount),
			InStockCount:    int(row.InStockCount),
		})
	}
	return out, nil
}

func text(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
