package ops

import (
	"encoding/json"
	"time"
)

type pharmacyDTO struct {
	ID         int64  `json:"id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	SiteDomain string `json:"site_domain"`
	LogoURL    string `json:"logo_url,omitempty"`
	Status     string `json:"status"`
}

type sourceDTO struct {
	ID               int64           `json:"id"`
	PharmacyID       int64           `json:"pharmacy_id"`
	PharmacyName     string          `json:"pharmacy_name"`
	PharmacySlug     string          `json:"pharmacy_slug"`
	Kind             string          `json:"kind"`
	Config           json.RawMessage `json:"config"`
	ScheduleInterval string          `json:"schedule_interval"`
	Enabled          bool            `json:"enabled"`
	LastRunAt        *time.Time      `json:"last_run_at,omitempty"`
	LastStatus       string          `json:"last_status,omitempty"`
	LastError        string          `json:"last_error,omitempty"`
}

type matchDTO struct {
	ID                 int64           `json:"id"`
	SourceItemID       int64           `json:"source_item_id"`
	SourceID           int64           `json:"source_id"`
	ExternalID         string          `json:"external_id"`
	Score              float64         `json:"score"`
	Status             string          `json:"status"`
	Reason             string          `json:"reason"`
	Raw                json.RawMessage `json:"raw"`
	CurrentProductID   int64           `json:"current_product_id,omitempty"`
	SuggestedProductID int64           `json:"suggested_product_id,omitempty"`
	SuggestedSlug      string          `json:"suggested_slug,omitempty"`
	SuggestedName      string          `json:"suggested_name,omitempty"`
	SuggestedImage     string          `json:"suggested_image,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
}

type productDTO struct {
	ID             int64           `json:"id"`
	Slug           string          `json:"slug"`
	NameFa         string          `json:"name_fa"`
	NameEn         string          `json:"name_en,omitempty"`
	GenericName    string          `json:"generic_name,omitempty"`
	BrandID        int64           `json:"brand_id,omitempty"`
	BrandName      string          `json:"brand_name,omitempty"`
	CategoryID     int64           `json:"category_id,omitempty"`
	CategoryName   string          `json:"category_name,omitempty"`
	ImageURL       string          `json:"image_url,omitempty"`
	Status         string          `json:"status"`
	LockedFields   []string        `json:"locked_fields"`
	SourceSnapshot json.RawMessage `json:"source_snapshot"`
}

type categoryDTO struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id,omitempty"`
	Slug     string `json:"slug"`
	NameFa   string `json:"name_fa"`
	Position int32  `json:"position"`
}

type brandDTO struct {
	ID     int64  `json:"id"`
	Slug   string `json:"slug"`
	NameFa string `json:"name_fa"`
	NameEn string `json:"name_en,omitempty"`
}

type auditDTO struct {
	ID        int64           `json:"id"`
	ActorID   int64           `json:"actor_id,omitempty"`
	ActorName string          `json:"actor_name"`
	Entity    string          `json:"entity"`
	EntityID  string          `json:"entity_id"`
	Action    string          `json:"action"`
	Before    json.RawMessage `json:"before"`
	After     json.RawMessage `json:"after"`
	CreatedAt time.Time       `json:"created_at"`
}

type jobDTO struct {
	JobID     int64     `json:"job_id,omitempty"`
	Kind      string    `json:"kind,omitempty"`
	Status    string    `json:"status,omitempty"`
	Attempts  int32     `json:"attempts,omitempty"`
	RunAt     time.Time `json:"run_at,omitempty"`
	LastError string    `json:"last_error,omitempty"`
	Enqueued  bool      `json:"enqueued"`
	LatestRun *runDTO   `json:"latest_run,omitempty"`
}

type runDTO struct {
	ID         int64      `json:"id"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	OkCount    int32      `json:"ok_count"`
	FailCount  int32      `json:"fail_count"`
	Error      string     `json:"error,omitempty"`
}

func toPharmacy(p Pharmacy) pharmacyDTO {
	return pharmacyDTO(p)
}

func toPharmacies(items []Pharmacy) []pharmacyDTO {
	out := make([]pharmacyDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toPharmacy(item))
	}
	return out
}

func toSource(s Source) sourceDTO {
	return sourceDTO{
		ID: s.ID, PharmacyID: s.PharmacyID, PharmacyName: s.PharmacyName, PharmacySlug: s.PharmacySlug,
		Kind: s.Kind, Config: rawJSON(s.Config), ScheduleInterval: s.ScheduleInterval, Enabled: s.Enabled,
		LastRunAt: s.LastRunAt, LastStatus: s.LastStatus, LastError: s.LastError,
	}
}

func toSources(items []Source) []sourceDTO {
	out := make([]sourceDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toSource(item))
	}
	return out
}

func toMatch(m Match) matchDTO {
	return matchDTO{
		ID: m.ID, SourceItemID: m.SourceItemID, SourceID: m.SourceID, ExternalID: m.ExternalID,
		Score: m.Score, Status: m.Status, Reason: m.Reason, Raw: rawJSON(m.Raw),
		CurrentProductID: m.CurrentProductID, SuggestedProductID: m.SuggestedProductID,
		SuggestedSlug: m.SuggestedSlug, SuggestedName: m.SuggestedName, SuggestedImage: m.SuggestedImage,
		CreatedAt: m.CreatedAt,
	}
}

func toMatches(items []Match) []matchDTO {
	out := make([]matchDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toMatch(item))
	}
	return out
}

func toProduct(p Product) productDTO {
	return productDTO{
		ID: p.ID, Slug: p.Slug, NameFa: p.NameFa, NameEn: p.NameEn, GenericName: p.GenericName,
		BrandID: p.BrandID, BrandName: p.BrandName, CategoryID: p.CategoryID, CategoryName: p.CategoryName,
		ImageURL: p.ImageURL, Status: p.Status, LockedFields: p.LockedFields, SourceSnapshot: rawJSON(p.SourceSnapshot),
	}
}

func toProducts(items []Product) []productDTO {
	out := make([]productDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toProduct(item))
	}
	return out
}

func toCategory(c Category) categoryDTO { return categoryDTO(c) }

func toCategories(items []Category) []categoryDTO {
	out := make([]categoryDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toCategory(item))
	}
	return out
}

func toBrand(b Brand) brandDTO { return brandDTO(b) }

func toBrands(items []Brand) []brandDTO {
	out := make([]brandDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toBrand(item))
	}
	return out
}

func toAudit(e AuditEntry) auditDTO {
	return auditDTO{
		ID: e.ID, ActorID: e.ActorID, ActorName: e.ActorName, Entity: e.Entity, EntityID: e.EntityID,
		Action: e.Action, Before: rawJSON(e.Before), After: rawJSON(e.After), CreatedAt: e.CreatedAt,
	}
}

func toAudits(items []AuditEntry) []auditDTO {
	out := make([]auditDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toAudit(item))
	}
	return out
}

func toJob(j JobStatus) jobDTO {
	var run *runDTO
	if j.LatestRun != nil {
		run = &runDTO{
			ID: j.LatestRun.ID, Status: j.LatestRun.Status, StartedAt: j.LatestRun.StartedAt,
			FinishedAt: j.LatestRun.FinishedAt, OkCount: j.LatestRun.OkCount, FailCount: j.LatestRun.FailCount,
			Error: j.LatestRun.Error,
		}
	}
	return jobDTO{
		JobID: j.JobID, Kind: j.Kind, Status: j.Status, Attempts: j.Attempts,
		RunAt: j.RunAt, LastError: j.LastError, Enqueued: j.Enqueued, LatestRun: run,
	}
}

type healthDTO struct {
	ID               int64      `json:"id"`
	PharmacyID       int64      `json:"pharmacy_id"`
	PharmacyName     string     `json:"pharmacy_name"`
	PharmacySlug     string     `json:"pharmacy_slug"`
	Kind             string     `json:"kind"`
	ScheduleInterval string     `json:"schedule_interval"`
	Enabled          bool       `json:"enabled"`
	LastRunAt        *time.Time `json:"last_run_at,omitempty"`
	LastStatus       string     `json:"last_status,omitempty"`
	LastError        string     `json:"last_error,omitempty"`
	Overdue          bool       `json:"overdue"`
	LowTrust         bool       `json:"low_trust"`
	RejectCount      int32      `json:"reject_count"`
	SuccessRate      float64    `json:"success_rate"`
	Runs             int32      `json:"runs"`
	Succeeded        int32      `json:"succeeded"`
	RunDurationMS    int64      `json:"run_duration_ms,omitempty"`
	LatestRun        *runDTO    `json:"latest_run,omitempty"`
}

type suspiciousDTO struct {
	ID                int64      `json:"id"`
	ProductID         int64      `json:"product_id"`
	ProductSlug       string     `json:"product_slug"`
	ProductName       string     `json:"product_name"`
	PharmacyID        int64      `json:"pharmacy_id"`
	PharmacyName      string     `json:"pharmacy_name"`
	PharmacySlug      string     `json:"pharmacy_slug"`
	SourceID          int64      `json:"source_id,omitempty"`
	SourceRejectCount int32      `json:"source_reject_count"`
	CurrentPriceRial  int64      `json:"current_price_rial"`
	ProposedPriceRial int64      `json:"proposed_price_rial"`
	ProposedAt        *time.Time `json:"proposed_at,omitempty"`
	ProductURL        string     `json:"product_url"`
	LastSeenAt        time.Time  `json:"last_seen_at"`
	InStock           bool       `json:"in_stock"`
}

type staleDTO struct {
	ID           int64     `json:"id"`
	PharmacyName string    `json:"pharmacy_name"`
	PharmacySlug string    `json:"pharmacy_slug"`
	OldestSeen   time.Time `json:"oldest_seen"`
	OfferCount   int32     `json:"offer_count"`
	AgeHours     float64   `json:"age_hours"`
}

type clickPharmacyDTO struct {
	ID     int64  `json:"id"`
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	Clicks int64  `json:"clicks"`
}

type clickProductDTO struct {
	ID     int64  `json:"id"`
	Slug   string `json:"slug"`
	NameFa string `json:"name_fa"`
	Clicks int64  `json:"clicks"`
}

func toHealth(h SourceHealth) healthDTO {
	var run *runDTO
	if h.LatestRun != nil {
		run = &runDTO{
			ID: h.LatestRun.ID, Status: h.LatestRun.Status, StartedAt: h.LatestRun.StartedAt,
			FinishedAt: h.LatestRun.FinishedAt, OkCount: h.LatestRun.OkCount, FailCount: h.LatestRun.FailCount,
			Error: h.LatestRun.Error,
		}
	}
	return healthDTO{
		ID: h.ID, PharmacyID: h.PharmacyID, PharmacyName: h.PharmacyName, PharmacySlug: h.PharmacySlug,
		Kind: h.Kind, ScheduleInterval: h.ScheduleInterval, Enabled: h.Enabled, LastRunAt: h.LastRunAt,
		LastStatus: h.LastStatus, LastError: h.LastError, Overdue: h.Overdue, LowTrust: h.LowTrust,
		RejectCount: h.RejectCount, SuccessRate: h.SuccessRate, Runs: h.Runs, Succeeded: h.Succeeded,
		RunDurationMS: h.RunDurationMS, LatestRun: run,
	}
}

func toHealthList(items []SourceHealth) []healthDTO {
	out := make([]healthDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toHealth(item))
	}
	return out
}

func toSuspicious(o SuspiciousOffer) suspiciousDTO {
	return suspiciousDTO{
		ID: o.ID, ProductID: o.ProductID, ProductSlug: o.ProductSlug, ProductName: o.ProductName,
		PharmacyID: o.PharmacyID, PharmacyName: o.PharmacyName, PharmacySlug: o.PharmacySlug,
		SourceID: o.SourceID, SourceRejectCount: o.SourceRejectCount,
		CurrentPriceRial: o.CurrentPriceRial, ProposedPriceRial: o.ProposedPriceRial,
		ProposedAt: o.ProposedAt, ProductURL: o.ProductURL, LastSeenAt: o.LastSeenAt, InStock: o.InStock,
	}
}

func toSuspiciousList(items []SuspiciousOffer) []suspiciousDTO {
	out := make([]suspiciousDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toSuspicious(item))
	}
	return out
}

func toStaleList(items []StaleSource) []staleDTO {
	out := make([]staleDTO, 0, len(items))
	for _, item := range items {
		out = append(out, staleDTO(item))
	}
	return out
}
