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
}
