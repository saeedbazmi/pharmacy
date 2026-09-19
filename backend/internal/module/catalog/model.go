package catalog

import (
	"errors"
	"time"
)

// Domain errors. The HTTP layer maps these to status codes in one place.
var (
	// ErrProductNotFound means no published product matches the identifier.
	ErrProductNotFound = errors.New("product not found")
	// ErrInvalidSlug means the caller supplied an unusable identifier.
	ErrInvalidSlug = errors.New("invalid product slug")
	// ErrCategoryNotFound means no live category matches the identifier.
	ErrCategoryNotFound = errors.New("category not found")
)

// Product is a unified catalog item: one row per real-world product, no matter
// how many pharmacies sell it. Absent optional values are empty strings.
type Product struct {
	ID           int64
	Slug         string
	NameFa       string
	NameEn       string
	GenericName  string
	DosageForm   string
	Strength     string
	ImageURL     string
	Description  string
	BrandName    string
	CategoryName string
	UpdatedAt    time.Time
	Offers       []Offer
}

// Offer is the public price of a product at one pharmacy.
type Offer struct {
	ID           int64
	PriceRial    int64
	InStock      bool
	ProductURL   string
	LastSeenAt   time.Time
	PharmacyName string
	PharmacySlug string
	BestPrice    bool
	Stale        bool
}

// SitemapEntry is a published slug for the public sitemap.
type SitemapEntry struct {
	Slug      string
	UpdatedAt time.Time
}

// ProductSummary is a compact row for lists and the compare table.
type ProductSummary struct {
	ID              int64
	Slug            string
	NameFa          string
	ImageURL        string
	BrandName       string
	LowestPriceRial int64
	OfferCount      int
	InStockCount    int
}

// Category is a published catalogue grouping used on public category pages.
type Category struct {
	ID        int64
	Slug      string
	NameFa    string
	UpdatedAt time.Time
}
