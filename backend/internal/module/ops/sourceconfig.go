package ops

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
)

var allowedSchedules = map[string]string{
	"15m":        "15 minutes",
	"15 minutes": "15 minutes",
	"30m":        "30 minutes",
	"30 minutes": "30 minutes",
	"1h":         "1 hour",
	"1 hour":     "1 hour",
	"2h":         "2 hours",
	"2 hours":    "2 hours",
	"6h":         "6 hours",
	"6 hours":    "6 hours",
	"12h":        "12 hours",
	"12 hours":   "12 hours",
	"24h":        "24 hours",
	"24 hours":   "24 hours",
	"1 day":      "24 hours",
}

type sourceConfig struct {
	Fetcher    string `json:"fetcher"`
	ListingURL string `json:"listing_url"`
	MaxPages   int    `json:"max_pages"`
	APIKey     string `json:"api_key,omitempty"`
}

func normalizeSchedule(raw string) (string, error) {
	key := strings.ToLower(strings.TrimSpace(raw))
	if key == "" {
		return "1 hour", nil
	}
	got, ok := allowedSchedules[key]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrInvalidSchedule, raw)
	}
	return got, nil
}

func parseSourceConfig(raw json.RawMessage, fetchers []string) (sourceConfig, error) {
	var cfg sourceConfig
	if err := json.Unmarshal(nonzeroJSON(raw), &cfg); err != nil {
		return cfg, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	cfg.Fetcher = strings.TrimSpace(strings.ToLower(cfg.Fetcher))
	if cfg.Fetcher == "" {
		return cfg, fmt.Errorf("%w: fetcher is required", ErrInvalidConfig)
	}
	if !contains(fetchers, cfg.Fetcher) {
		return cfg, fmt.Errorf("%w: %s", ErrUnknownFetcher, cfg.Fetcher)
	}
	cfg.ListingURL = strings.TrimSpace(cfg.ListingURL)
	if cfg.ListingURL == "" {
		return cfg, fmt.Errorf("%w: listing_url is required", ErrInvalidConfig)
	}
	if err := crawlhttp.ValidateURL(cfg.ListingURL); err != nil {
		return cfg, fmt.Errorf("%w: %v", ErrUnsafeURL, err)
	}
	if cfg.MaxPages <= 0 {
		cfg.MaxPages = 1
	}
	if cfg.MaxPages > 1 {
		cfg.MaxPages = 1
	}
	return cfg, nil
}

func normalizeKind(kind string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "api", "crawler":
		return strings.ToLower(kind), nil
	default:
		return "", fmt.Errorf("%w: kind must be api or crawler", ErrInvalidInput)
	}
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}

func allowedRole(role string) bool {
	return role == RoleDataOps || role == RoleSuperAdmin
}
