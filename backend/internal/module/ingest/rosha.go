package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
)

type roshaConfig struct {
	ListingURL string `json:"listing_url"`
	MaxPages   int    `json:"max_pages"`
}

// RoshaFetcher crawls one Rosha category listing. M2 keeps max_pages at 1.
type RoshaFetcher struct {
	http *crawlhttp.Client
}

func NewRoshaFetcher(httpClient *crawlhttp.Client) *RoshaFetcher {
	return &RoshaFetcher{http: httpClient}
}

func (f *RoshaFetcher) Fetch(ctx context.Context, src Source) ([]RawItem, error) {
	cfg, err := parseRoshaConfig(src.Config)
	if err != nil {
		return nil, err
	}
	if err := crawlhttp.ValidateURL(cfg.ListingURL); err != nil {
		return nil, fmt.Errorf("listing url: %w", err)
	}
	body, err := f.http.Get(ctx, cfg.ListingURL)
	if err != nil {
		return nil, fmt.Errorf("fetch rosha listing: %w", err)
	}
	items, err := ParseRoshaListing(body)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func parseRoshaConfig(raw json.RawMessage) (roshaConfig, error) {
	var cfg roshaConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("rosha config: %w", err)
	}
	if cfg.ListingURL == "" {
		return cfg, fmt.Errorf("rosha config missing listing_url")
	}
	cfg.MaxPages = 1
	return cfg, nil
}
