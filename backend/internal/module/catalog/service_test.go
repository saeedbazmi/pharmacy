package catalog

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
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

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

func TestProductBySlugReturnsProduct(t *testing.T) {
	repo := &fakeProductReader{product: Product{ID: 7, Slug: "acetaminophen-500", NameFa: "استامینوفن ۵۰۰"}}
	svc := NewService(repo, discardLogger())

	got, err := svc.ProductBySlug(context.Background(), "  Acetaminophen-500  ")
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

			_, err := svc.ProductBySlug(context.Background(), slug)
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

	_, err := svc.ProductBySlug(context.Background(), "missing-product")
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("error = %v, want ErrProductNotFound to survive wrapping", err)
	}
}
