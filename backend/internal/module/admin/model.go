package admin

import (
	"encoding/json"
	"time"
)

const (
	RoleDataOps        = "data_ops"
	RoleSuperAdmin     = "super_admin"
	FlagDirectPurchase = "direct_purchase"

	settingSEO      = "seo"
	settingBanner   = "banner"
	settingFeatured = "featured_categories"
	settingPages    = "pages"

	flagCacheKey     = "flags"
	settingsCacheKey = "settings"
	maxRangeDays     = 366
	defaultRangeDays = 7
	minPasswordLen   = 10
)

// Actor is an authenticated internal operator reaching /admin.
type Actor struct {
	ID       int64
	Username string
	Role     string
}

// Coverage is catalogue health shown on the dashboard.
type Coverage struct {
	ActivePharmacies       int64
	PublishedProducts      int64
	MedianFreshnessSeconds float64
}

// SearchTerm is an aggregated query for the dashboard lists.
type SearchTerm struct {
	Query    string
	Hits     int64
	LastSeen time.Time
}

// PharmacyClicks is redirect volume for one pharmacy.
type PharmacyClicks struct {
	ID     int64
	Slug   string
	Name   string
	Clicks int64
}

// DayPoint is one calendar day of traffic.
type DayPoint struct {
	Day      time.Time
	Searches int64
	Clicks   int64
}

// Dashboard is the super_admin report for a closed time range.
type Dashboard struct {
	From       time.Time
	To         time.Time
	Days       int
	Searches   int64
	Clicks     int64
	CTR        float64
	Coverage   Coverage
	Pharmacies []PharmacyClicks
	Frequent   []SearchTerm
	Zero       []SearchTerm
	Series     []DayPoint
}

// InternalUser is a panel login, never including the password hash.
type InternalUser struct {
	ID        int64
	Username  string
	Role      string
	IsActive  bool
	CreatedAt time.Time
}

// UserPatch is a role, active-flag or password change.
type UserPatch struct {
	Role     string
	IsActive *bool
	Password string
}

// Flag is one feature-flag row; PharmacyID 0 means global.
type Flag struct {
	Key        string
	PharmacyID int64
	Enabled    bool
	UpdatedAt  time.Time
}

// PharmacyBrief is enough to label a per-pharmacy flag.
type PharmacyBrief struct {
	ID   int64
	Slug string
	Name string
}

// SEOSettings are the default public title and description.
type SEOSettings struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// BannerSettings is the optional home hero.
type BannerSettings struct {
	Enabled bool   `json:"enabled"`
	Title   string `json:"title"`
	Body    string `json:"body"`
}

// FeaturedSettings lists category slugs highlighted on the home page.
type FeaturedSettings struct {
	Slugs []string `json:"slugs"`
}

// PageSettings are the legal and about texts.
type PageSettings struct {
	About      string `json:"about"`
	Contact    string `json:"contact"`
	Terms      string `json:"terms"`
	Disclaimer string `json:"disclaimer"`
}

// Settings is the public and admin view of site_settings.
type Settings struct {
	SEO      SEOSettings      `json:"seo"`
	Banner   BannerSettings   `json:"banner"`
	Featured FeaturedSettings `json:"featured_categories"`
	Pages    PageSettings     `json:"pages"`
}

func defaultSettings() Settings {
	return Settings{
		SEO: SEOSettings{
			Title:       "مقایسه قیمت دارو",
			Description: "قیمت دارو و محصولات سلامت را در چند داروخانه آنلاین مقایسه کنید و برای خرید به سایت همان داروخانه بروید.",
		},
		Banner:   BannerSettings{},
		Featured: FeaturedSettings{Slugs: []string{}},
		Pages: PageSettings{
			About:      "این وب‌سایت قیمت دارو و محصولات سلامت را از داروخانه‌های آنلاین جمع می‌کند تا بتوانید مقایسه کنید و برای خرید به سایت همان داروخانه بروید. ما فروشنده نیستیم و سفارشی ثبت نمی‌کنیم.",
			Contact:    "برای گزارش خطای داده یا درخواست منبع جدید، از طریق ایمیل عملیاتی اعلام‌شده در استقرار پیام بفرستید. این نشانی برای مشاوره درمانی نیست.",
			Terms:      "استفاده از این وب‌سایت به معنای پذیرش این قواعد است: اطلاعات قیمت از منابع داروخانه‌ها می‌آید و ممکن است تا چند ساعت کهنه باشد. خرید و پرداخت فقط در سایت داروخانه انجام می‌شود. پلتفرم واسطه فروش، سبد خرید یا پرداخت نیست.",
			Disclaimer: "این وب‌سایت مرجع تشخیص یا درمان نیست و توصیه پزشکی یا دارویی ارائه نمی‌دهد. تصمیم درمانی را با پزشک یا داروساز بگیرید. قیمت و موجودی نهایی فقط در سایت داروخانه معتبر است.",
		},
	}
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}
