package redirect

import (
	"errors"
	"time"
)

var (
	// ErrOfferNotFound means the buy link is stale or never existed.
	ErrOfferNotFound = errors.New("offer not found")
	// ErrInvalidOfferID means the path value is not a positive integer.
	ErrInvalidOfferID = errors.New("invalid offer id")
)

// Target is the destination of a buy click: only the stored product URL.
type Target struct {
	OfferID    int64
	ProductID  int64
	PharmacyID int64
	ProductURL string
}

// Click is one recorded redirect, written off the request path.
type Click struct {
	OfferID    int64
	ProductID  int64
	PharmacyID int64
	Referrer   string
	ClickedAt  time.Time
}
