package search

import (
	"context"
	"fmt"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

type postgresPort struct {
	q          *store.Queries
	freshSince func() time.Time
}

func NewPostgres(db store.DBTX) Port {
	return NewPostgresWithFreshness(db, 72*time.Hour)
}

func NewPostgresWithFreshness(db store.DBTX, criticalAfter time.Duration) Port {
	if criticalAfter <= 0 {
		criticalAfter = 72 * time.Hour
	}
	return &postgresPort{
		q: store.New(db),
		freshSince: func() time.Time {
			return time.Now().Add(-criticalAfter)
		},
	}
}

type offerHit struct {
	ID              int64
	Slug            string
	NameFa          string
	ImageUrl        *string
	BrandName       *string
	LowestPriceRial int64
	OfferCount      int32
	InStockCount    int32
}

func (p *postgresPort) Search(ctx context.Context, q Query) (Result, error) {
	fresh := p.freshSince()
	total, err := p.q.CountSearchProducts(ctx, store.CountSearchProductsParams{
		SimLimit:     float32(MinSimilarity),
		Q:            q.Q,
		BrandSlug:    q.Brand,
		CategorySlug: q.Category,
		InStockOnly:  q.InStock,
		FreshSince:   fresh,
	})
	if err != nil {
		return Result{}, fmt.Errorf("count search products: %w", err)
	}
	out := Result{Hits: make([]Hit, 0), Total: int(total), Page: q.Page, PageSize: q.PageSize}
	if total == 0 {
		return out, nil
	}
	var rows []offerHit
	if q.Sort == SortPriceAsc || q.Sort == SortPriceDesc {
		priced, err := p.q.SearchProductsByPrice(ctx, store.SearchProductsByPriceParams{
			SimLimit:     float32(MinSimilarity),
			Q:            q.Q,
			BrandSlug:    q.Brand,
			CategorySlug: q.Category,
			InStockOnly:  q.InStock,
			Sort:         q.Sort,
			PageLimit:    int32(q.PageSize),
			PageOffset:   int32((q.Page - 1) * q.PageSize),
			FreshSince:   fresh,
		})
		if err != nil {
			return Result{}, fmt.Errorf("search products by price: %w", err)
		}
		rows = make([]offerHit, 0, len(priced))
		for _, row := range priced {
			rows = append(rows, offerHit{
				ID:              row.ID,
				Slug:            row.Slug,
				NameFa:          row.NameFa,
				ImageUrl:        row.ImageUrl,
				BrandName:       row.BrandName,
				LowestPriceRial: row.LowestPriceRial,
				OfferCount:      row.OfferCount,
				InStockCount:    row.InStockCount,
			})
		}
	} else if total <= knnMinMatches {
		matched, err := p.q.SearchProductsMatched(ctx, store.SearchProductsMatchedParams{
			SimLimit:     float32(MinSimilarity),
			Q:            q.Q,
			BrandSlug:    q.Brand,
			CategorySlug: q.Category,
			InStockOnly:  q.InStock,
			PageLimit:    int32(q.PageSize),
			PageOffset:   int32((q.Page - 1) * q.PageSize),
			FreshSince:   fresh,
		})
		if err != nil {
			return Result{}, fmt.Errorf("search products matched: %w", err)
		}
		rows = make([]offerHit, 0, len(matched))
		for _, row := range matched {
			rows = append(rows, offerHit{
				ID:              row.ID,
				Slug:            row.Slug,
				NameFa:          row.NameFa,
				ImageUrl:        row.ImageUrl,
				BrandName:       row.BrandName,
				LowestPriceRial: row.LowestPriceRial,
				OfferCount:      row.OfferCount,
				InStockCount:    row.InStockCount,
			})
		}
	} else {
		ranked, err := p.q.SearchProducts(ctx, store.SearchProductsParams{
			Q:            q.Q,
			BrandSlug:    q.Brand,
			CategorySlug: q.Category,
			InStockOnly:  q.InStock,
			PageLimit:    int32(q.PageSize),
			PageOffset:   int32((q.Page - 1) * q.PageSize),
			FreshSince:   fresh,
		})
		if err != nil {
			return Result{}, fmt.Errorf("search products: %w", err)
		}
		rows = make([]offerHit, 0, len(ranked))
		for _, row := range ranked {
			if row.Sim < MinSimilarity {
				continue
			}
			rows = append(rows, offerHit{
				ID:              row.ID,
				Slug:            row.Slug,
				NameFa:          row.NameFa,
				ImageUrl:        row.ImageUrl,
				BrandName:       row.BrandName,
				LowestPriceRial: row.LowestPriceRial,
				OfferCount:      row.OfferCount,
				InStockCount:    row.InStockCount,
			})
		}
	}
	out.Hits = make([]Hit, 0, len(rows))
	for _, row := range rows {
		out.Hits = append(out.Hits, Hit{
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

func (p *postgresPort) Facets(ctx context.Context) (brands, categories []Facet, err error) {
	brandRows, err := p.q.ListSearchBrands(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list search brands: %w", err)
	}
	catRows, err := p.q.ListSearchCategories(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("list search categories: %w", err)
	}
	brands = make([]Facet, 0, len(brandRows))
	for _, row := range brandRows {
		brands = append(brands, Facet{Slug: row.Slug, Name: row.NameFa})
	}
	categories = make([]Facet, 0, len(catRows))
	for _, row := range catRows {
		categories = append(categories, Facet{Slug: row.Slug, Name: row.NameFa})
	}
	return brands, categories, nil
}

func text(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
