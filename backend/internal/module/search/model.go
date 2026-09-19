package search

import (
	"context"
	"errors"
	"time"
)

var (
	ErrQueryEmpty   = errors.New("search query empty")
	ErrQueryTooLong = errors.New("search query too long")
)

const (
	SortRelevance = "relevance"
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"
	MinSimilarity = 0.28
	DefaultPage   = 1
	DefaultSize   = 24
	MaxPageSize   = 48
	// knnMinMatches: below this, GIN-filtered ranking is used so a rare
	// token like a brand name is not lost among 30k near-identical rows.
	knnMinMatches = 512
)

// Query is the public search request after transport validation.
type Query struct {
	Q        string
	Brand    string
	Category string
	InStock  bool
	Sort     string
	Page     int
	PageSize int
}

// Hit is one published product in a result page.
type Hit struct {
	ID              int64
	Slug            string
	NameFa          string
	ImageURL        string
	BrandName       string
	LowestPriceRial int64
	OfferCount      int
	InStockCount    int
}

// Facet is a brand or category the UI can filter on.
type Facet struct {
	Slug string
	Name string
}

// Result is one cached page of search hits.
type Result struct {
	Hits       []Hit
	Total      int
	Page       int
	PageSize   int
	Brands     []Facet
	Categories []Facet
}

// Port is the search backend. Postgres is the phase-1 adapter; Meilisearch can
// replace it later without changing the service.
type Port interface {
	Search(ctx context.Context, q Query) (Result, error)
	Facets(ctx context.Context) (brands, categories []Facet, err error)
}

// Event is one recorded search for reporting.
type Event struct {
	Query       string
	ResultCount int
	Duration    time.Duration
	UserID      *int64
}
