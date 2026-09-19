package ingest

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
)

type darukadeConfig struct {
	ListingURL string `json:"listing_url"`
	MaxPages   int    `json:"max_pages"`
}

// DarukadeFetcher crawls one Darukade category listing. M1 is limited to a
// single page via max_pages=1 in data_sources.config.
type DarukadeFetcher struct {
	http *crawlhttp.Client
}

func NewDarukadeFetcher(httpClient *crawlhttp.Client) *DarukadeFetcher {
	return &DarukadeFetcher{http: httpClient}
}

func (f *DarukadeFetcher) Fetch(ctx context.Context, src Source) ([]RawItem, error) {
	cfg, err := parseDarukadeConfig(src.Config)
	if err != nil {
		return nil, err
	}
	if err := crawlhttp.ValidateURL(cfg.ListingURL); err != nil {
		return nil, fmt.Errorf("listing url: %w", err)
	}
	body, err := f.http.Get(ctx, cfg.ListingURL)
	if err != nil {
		return nil, fmt.Errorf("fetch darukade listing: %w", err)
	}
	items, err := ParseListing(body, "https://darukade.com")
	if err != nil {
		return nil, err
	}
	return items, nil
}

func parseDarukadeConfig(raw json.RawMessage) (darukadeConfig, error) {
	var cfg darukadeConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("darukade config: %w", err)
	}
	if cfg.ListingURL == "" {
		return cfg, fmt.Errorf("darukade config missing listing_url")
	}
	if cfg.MaxPages == 0 {
		cfg.MaxPages = 1
	}
	if cfg.MaxPages != 1 {
		// M1 scope: never walk the rest of the catalogue, even if misconfigured.
		cfg.MaxPages = 1
	}
	return cfg, nil
}

func fetcherName(cfg json.RawMessage) (string, error) {
	var wrap struct {
		Fetcher string `json:"fetcher"`
	}
	if err := json.Unmarshal(cfg, &wrap); err != nil {
		return "", fmt.Errorf("source config: %w", err)
	}
	if wrap.Fetcher == "" {
		return "", fmt.Errorf("source config missing fetcher")
	}
	return wrap.Fetcher, nil
}
