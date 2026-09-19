package alerting

import (
	"errors"
	"time"
)

const (
	KindPriceDrop   = "price_drop"
	KindBackInStock = "back_in_stock"
	MaxActiveAlerts = 20
	stockEventKey   = "stock:in"
)

var (
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidInput     = errors.New("invalid input")
	ErrProductNotFound  = errors.New("product not found")
	ErrAlertLimit       = errors.New("alert limit reached")
	ErrNotFound         = errors.New("not found")
	ErrInvalidAlertKind = errors.New("invalid alert kind")
)

// Favorite is a saved product with the current lowest fresh price.
type Favorite struct {
	ProductID       int64
	Slug            string
	NameFa          string
	ImageURL        string
	LowestPriceRial int64
	OfferCount      int
	CreatedAt       time.Time
}

// Alert is a price or stock watch the user asked for.
type Alert struct {
	ID                int64
	UserID            int64
	ProductID         int64
	Kind              string
	TargetPriceRial   *int64
	BaselinePriceRial int64
	StockWasAvailable bool
	Status            string
	CreatedAt         time.Time
	ProductSlug       string
	ProductName       string
}

// AlertInput is the create payload after transport validation.
type AlertInput struct {
	ProductID       int64
	Kind            string
	TargetPriceRial *int64
}

// Quote is the current valid (not suspicious, not stale) market for a product.
type Quote struct {
	ProductID         int64
	LowestInStockRial int64
	InStock           bool
}

type evalAlert struct {
	Alert
	Phone       string
	ProductName string
}

type delivery struct {
	ID          int64
	AlertID     int64
	EventKey    string
	Attempts    int32
	UserID      int64
	ProductID   int64
	Kind        string
	Phone       string
	ProductName string
}
