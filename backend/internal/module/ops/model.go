package ops

import (
	"encoding/json"
	"time"
)

const (
	RoleDataOps         = "data_ops"
	RoleSuperAdmin      = "super_admin"
	SessionCookie       = "ops_session"
	sessionTTL          = 12 * time.Hour
	loginWindow         = 15 * time.Minute
	loginMaxAttempt     = 8
	previewLimit        = 8
	lowTrustRejects     = 3
	defaultHealthWindow = 7 * 24 * time.Hour
	defaultClickWindow  = 30 * 24 * time.Hour
)

// Actor is an authenticated internal operator.
type Actor struct {
	ID       int64
	Username string
	Role     string
}

// Session is a live login.
type Session struct {
	Token     string
	ExpiresAt time.Time
	Actor     Actor
}

// Pharmacy is a seller the panel manages.
type Pharmacy struct {
	ID         int64
	Slug       string
	Name       string
	SiteDomain string
	LogoURL    string
	Status     string
}

// Source is a data_sources row as the panel sees it. Config is always masked.
type Source struct {
	ID               int64
	PharmacyID       int64
	PharmacyName     string
	PharmacySlug     string
	Kind             string
	Config           json.RawMessage
	ScheduleInterval string
	Enabled          bool
	LastRunAt        *time.Time
	LastStatus       string
	LastError        string
}

// SourceInput is the writable shape of a source.
type SourceInput struct {
	PharmacyID       int64
	Kind             string
	Config           json.RawMessage
	ScheduleInterval string
	Enabled          bool
}

// PharmacyInput is the writable shape of a pharmacy.
type PharmacyInput struct {
	Slug       string
	Name       string
	SiteDomain string
	LogoURL    string
	Status     string
}

// PreviewItem is one fetched row shown by test-connection, never persisted.
type PreviewItem struct {
	ExternalID string
	NameFa     string
	BrandName  string
	PriceToman int64
	InStock    bool
	ProductURL string
	ImageURL   string
}

// JobStatus is a recent job or the latest sync run.
type JobStatus struct {
	JobID     int64
	Kind      string
	Status    string
	Attempts  int32
	RunAt     time.Time
	LastError string
	CreatedAt time.Time
	Enqueued  bool
	LatestRun *SyncRun
}

// SyncRun is the last crawl of a source.
type SyncRun struct {
	ID         int64
	Status     string
	StartedAt  time.Time
	FinishedAt *time.Time
	OkCount    int32
	FailCount  int32
	Error      string
}

// Match is one pending (or decided) candidate.
type Match struct {
	ID                 int64
	SourceItemID       int64
	SourceID           int64
	ExternalID         string
	Score              float64
	Status             string
	Reason             string
	Raw                json.RawMessage
	CurrentProductID   int64
	SuggestedProductID int64
	SuggestedSlug      string
	SuggestedName      string
	SuggestedImage     string
	CreatedAt          time.Time
}

// Product is the ops view of a catalog row, including hidden ones.
type Product struct {
	ID             int64
	Slug           string
	NameFa         string
	NameEn         string
	GenericName    string
	BrandID        int64
	BrandName      string
	CategoryID     int64
	CategoryName   string
	ImageURL       string
	Status         string
	LockedFields   []string
	SourceSnapshot json.RawMessage
}

// ProductPatch is an operator edit. Nil pointer means "leave unchanged".
type ProductPatch struct {
	NameFa      *string
	NameEn      *string
	GenericName *string
	BrandID     *int64
	CategoryID  *int64
	ImageURL    *string
	Status      *string
}

// Category is one node in the tree.
type Category struct {
	ID       int64
	ParentID int64
	Slug     string
	NameFa   string
	Position int32
}

// Brand is a catalogue brand.
type Brand struct {
	ID     int64
	Slug   string
	NameFa string
	NameEn string
}

// AuditEntry is one panel write.
type AuditEntry struct {
	ID        int64
	ActorID   int64
	ActorName string
	Entity    string
	EntityID  string
	Action    string
	Before    json.RawMessage
	After     json.RawMessage
	CreatedAt time.Time
}

// SuspiciousOffer is a withheld price jump waiting for a human.
type SuspiciousOffer struct {
	Status            string
	ID                int64
	ProductID         int64
	ProductSlug       string
	ProductName       string
	PharmacyID        int64
	PharmacyName      string
	PharmacySlug      string
	SourceID          int64
	SourceRejectCount int32
	CurrentPriceRial  int64
	ProposedPriceRial int64
	ProposedAt        *time.Time
	ProductURL        string
	LastSeenAt        time.Time
	InStock           bool
}

// SourceHealth is one row of the sync health table.
type SourceHealth struct {
	ID               int64
	PharmacyID       int64
	PharmacyName     string
	PharmacySlug     string
	Kind             string
	ScheduleInterval string
	Enabled          bool
	LastRunAt        *time.Time
	LastStatus       string
	LastError        string
	Overdue          bool
	LowTrust         bool
	RejectCount      int32
	SuccessRate      float64
	Runs             int32
	Succeeded        int32
	LatestRun        *SyncRun
	RunDurationMS    int64
}

// StaleSource is the oldest active offer age for a pharmacy source.
type StaleSource struct {
	ID           int64
	PharmacyName string
	PharmacySlug string
	OldestSeen   time.Time
	OfferCount   int32
	AgeHours     float64
}

// ClickPharmacy is click volume for one seller.
type ClickPharmacy struct {
	ID     int64
	Slug   string
	Name   string
	Clicks int64
}

// ClickProduct is click volume for one catalogue item.
type ClickProduct struct {
	ID     int64
	Slug   string
	NameFa string
	Clicks int64
}

// ClickReport is the phase-2 pharmacy-pilot input.
type ClickReport struct {
	Since       time.Time
	Until       time.Time
	Total       int64
	Pharmacies  []ClickPharmacy
	TopProducts []ClickProduct
	TopPharmacy *ClickPharmacy
}

// Page is a numbered slice.
type Page[T any] struct {
	Items []T
	Total int64
	Page  int
	Size  int
}
