package ingest

import (
	"context"
	"encoding/json"
	"time"
)

// Fetcher pulls a batch of raw items from one data source.
type Fetcher interface {
	Fetch(ctx context.Context, src Source) ([]RawItem, error)
}

// Source is the ingest view of a data_sources row.
type Source struct {
	ID         int64
	PharmacyID int64
	Kind       string
	Config     json.RawMessage
}

// RawItem is one product as the source returned it, before matching.
type RawItem struct {
	ExternalID string
	NameFa     string
	NameEn     string
	BrandName  string
	ImageURL   string
	ProductURL string
	PriceToman int64
	InStock    bool
	GTIN       string
	IRC        string
	SlugHint   string
	FetchedAt  time.Time
	Raw        json.RawMessage
}

// PriceRial converts the source toman price into rial for storage.
func (i RawItem) PriceRial() int64 {
	if i.PriceToman < 0 {
		return 0
	}
	return i.PriceToman * 10
}

// Registry maps the "fetcher" key in data_sources.config to an implementation.
// Worker wiring registers fetchers; ingest core never switches on pharmacy name.
type Registry struct {
	byName map[string]Fetcher
}

func NewRegistry() *Registry {
	return &Registry{byName: make(map[string]Fetcher)}
}

func (r *Registry) Register(name string, f Fetcher) {
	r.byName[name] = f
}

func (r *Registry) Get(name string) (Fetcher, bool) {
	f, ok := r.byName[name]
	return f, ok
}

// Names returns registered fetcher keys in insertion order.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.byName))
	for name := range r.byName {
		out = append(out, name)
	}
	return out
}
