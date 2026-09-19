package catalog

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeProductReader struct {
	product   Product
	err       error
	gotSlug   string
	callCount int
}

func (f *fakeProductReader) ProductBySlug(_ context.Context, slug string) (Product, error) {
	f.callCount++
	f.gotSlug = slug
	return f.product, f.err
}

func (f *fakeProductReader) ListPublished(context.Context, int32) ([]ProductSummary, error) {
	return nil, nil
}

func (f *fakeProductReader) ProductsBySlugs(context.Context, []string) ([]ProductSummary, error) {
	return nil, nil
}

func (f *fakeProductReader) ListSitemap(context.Context) ([]SitemapEntry, error) {
	return nil, nil
}

func (f *fakeProductReader) ListCategorySitemap(context.Context) ([]SitemapEntry, error) {
	return nil, nil
}

func (f *fakeProductReader) CategoryBySlug(context.Context, string) (Category, error) {
	return Category{}, ErrCategoryNotFound
}

func (f *fakeProductReader) ListByCategory(context.Context, string, int32) ([]ProductSummary, error) {
	return nil, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestProductBySlugReturnsProduct(t *testing.T) {
	repo := &fakeProductReader{product: Product{ID: 7, Slug: "acetaminophen-500", NameFa: "استامینوفن ۵۰۰"}}
	svc := NewService(repo, discardLogger())

	got, err := svc.ProductBySlug(context.Background(), "  Acetaminophen-500  ", OfferFilter{})
	if err != nil {
		t.Fatalf("ProductBySlug() returned error: %v", err)
	}
	if got.ID != 7 {
		t.Errorf("ID = %d, want 7", got.ID)
	}
	if repo.gotSlug != "acetaminophen-500" {
		t.Errorf("repository received %q, want normalized slug", repo.gotSlug)
	}
}

func TestProductBySlugRejectsBadSlugWithoutHittingRepository(t *testing.T) {
	tests := map[string]string{
		"empty":             "",
		"only spaces":       "   ",
		"persian letters":   "استامینوفن",
		"path traversal":    "../../etc/passwd",
		"sql injection try": "a' OR 1=1--",
		"too long":          string(make([]byte, 201)),
	}

	for name, slug := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &fakeProductReader{}
			svc := NewService(repo, discardLogger())

			_, err := svc.ProductBySlug(context.Background(), slug, OfferFilter{})
			if !errors.Is(err, ErrInvalidSlug) {
				t.Fatalf("error = %v, want ErrInvalidSlug", err)
			}
			if repo.callCount != 0 {
				t.Errorf("repository was called %d times, want 0", repo.callCount)
			}
		})
	}
}

func TestProductBySlugPropagatesNotFound(t *testing.T) {
	repo := &fakeProductReader{err: ErrProductNotFound}
	svc := NewService(repo, discardLogger())

	_, err := svc.ProductBySlug(context.Background(), "missing-product", OfferFilter{})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("error = %v, want ErrProductNotFound to survive wrapping", err)
	}
}

func TestRankByPriceMarksCheapestAndSorts(t *testing.T) {
	got := rankByPrice([]Offer{
		{ID: 1, PriceRial: 2000, PharmacyName: "ب"},
		{ID: 2, PriceRial: 1000, PharmacyName: "آ"},
		{ID: 3, PriceRial: 3000, PharmacyName: "پ"},
	})
	if got[0].ID != 2 || !got[0].BestPrice || got[1].BestPrice || got[2].BestPrice {
		t.Fatalf("ranked = %+v", got)
	}
}

func TestParseOfferFilterFallsBackOnInvalid(t *testing.T) {
	f := ParseOfferFilter(map[string][]string{
		"sort":     {"nope"},
		"in_stock": {"maybe"},
		"pharmacy": {"../x"},
	})
	if f.Sort != SortPriceAsc || f.InStock || f.Pharmacy != "" {
		t.Fatalf("%+v", f)
	}
}

func TestOfferFilterHidesOutOfStock(t *testing.T) {
	repo := &fakeProductReader{product: Product{
		Slug:      "x",
		BrandName: "Urelin",
		Offers: []Offer{
			{ID: 1, PriceRial: 100, InStock: false, PharmacySlug: "darukade", LastSeenAt: time.Now()},
			{ID: 2, PriceRial: 200, InStock: true, PharmacySlug: "rosha", LastSeenAt: time.Now()},
		},
	}}
	svc := NewService(repo, discardLogger())
	got, err := svc.ProductBySlug(context.Background(), "x", OfferFilter{InStock: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Offers) != 1 || got.Offers[0].ID != 2 || !got.Offers[0].BestPrice {
		t.Fatalf("%+v", got.Offers)
	}
}

func TestOfferFilterMarksStaleAndDropsCritical(t *testing.T) {
	now := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)
	repo := &fakeProductReader{product: Product{
		Slug: "x",
		Offers: []Offer{
			{ID: 1, PriceRial: 100, InStock: true, LastSeenAt: now.Add(-30 * time.Hour), PharmacyName: "کهنه"},
			{ID: 2, PriceRial: 90, InStock: true, LastSeenAt: now.Add(-80 * time.Hour), PharmacyName: "بحرانی"},
			{ID: 3, PriceRial: 110, InStock: true, LastSeenAt: now.Add(-2 * time.Hour), PharmacyName: "تازه"},
		},
	}}
	svc := NewService(repo, discardLogger())
	svc.now = func() time.Time { return now }

	got, err := svc.ProductBySlug(context.Background(), "x", OfferFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Offers) != 2 {
		t.Fatalf("offers = %d, want 2 (critical dropped)", len(got.Offers))
	}
	if got.Offers[0].ID != 1 || !got.Offers[0].Stale || !got.Offers[0].BestPrice {
		t.Fatalf("cheapest remaining = %+v", got.Offers[0])
	}
	if got.Offers[1].ID != 3 || got.Offers[1].Stale {
		t.Fatalf("fresh offer = %+v", got.Offers[1])
	}
}
