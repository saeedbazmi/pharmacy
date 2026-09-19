package catalog

import (
	"net/url"
	"strings"
	"time"
)

const (
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"
	// DefaultStaleAfter is the age at which a price is labelled as possibly outdated.
	DefaultStaleAfter = 24 * time.Hour
	// DefaultCriticalAfter is the age at which a price is dropped from comparison.
	DefaultCriticalAfter = 72 * time.Hour
	// StaleAfter is kept so existing callers and tests still compile.
	StaleAfter = DefaultStaleAfter
)

// OfferFilter is the public query for the product price table.
type OfferFilter struct {
	Sort     string
	InStock  bool
	Pharmacy string
	Brand    string
}

// ParseOfferFilter maps query params to a filter. Invalid values become defaults.
// Pharmacy/in-stock filters run in memory after ListActiveOffersByProduct, which
// already uses offers_product_price_idx. The product lookup uses products_slug_key.
func ParseOfferFilter(q url.Values) OfferFilter {
	f := OfferFilter{Sort: SortPriceAsc}
	switch q.Get("sort") {
	case SortPriceDesc:
		f.Sort = SortPriceDesc
	default:
		f.Sort = SortPriceAsc
	}
	switch q.Get("in_stock") {
	case "1", "true", "yes":
		f.InStock = true
	}
	f.Pharmacy = safeToken(q.Get("pharmacy"))
	f.Brand = strings.TrimSpace(q.Get("brand"))
	if len(f.Brand) > 80 {
		f.Brand = ""
	}
	return f
}

func applyOfferFilter(offers []Offer, f OfferFilter, now time.Time, staleAfter, criticalAfter time.Duration) []Offer {
	if staleAfter <= 0 {
		staleAfter = DefaultStaleAfter
	}
	if criticalAfter <= 0 {
		criticalAfter = DefaultCriticalAfter
	}
	out := make([]Offer, 0, len(offers))
	for _, o := range offers {
		age := now.Sub(o.LastSeenAt)
		if age > criticalAfter {
			continue
		}
		o.Stale = age > staleAfter
		if f.InStock && !o.InStock {
			continue
		}
		if f.Pharmacy != "" && o.PharmacySlug != f.Pharmacy {
			continue
		}
		out = append(out, o)
	}
	return out
}

func safeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || len(s) > 80 {
		return ""
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return ""
		}
	}
	return s
}
