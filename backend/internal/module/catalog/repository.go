package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

// ProductReader is the read side the service depends on. Declared here, on the
// consumer side, so the service never learns about pgx or generated code.
type ProductReader interface {
	ProductBySlug(ctx context.Context, slug string) (Product, error)
}

// Repository reads catalog rows through the generated queries.
type Repository struct {
	q *store.Queries
}

// NewRepository wires the repository to a pool, connection or transaction.
func NewRepository(db store.DBTX) *Repository {
	return &Repository{q: store.New(db)}
}

// ProductBySlug returns the published product with the given slug.
func (r *Repository) ProductBySlug(ctx context.Context, slug string) (Product, error) {
	row, err := r.q.GetProductBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, fmt.Errorf("query product by slug %q: %w", slug, err)
	}

	return Product{
		ID:           row.ID,
		Slug:         row.Slug,
		NameFa:       row.NameFa,
		NameEn:       text(row.NameEn),
		GenericName:  text(row.GenericName),
		DosageForm:   text(row.DosageForm),
		Strength:     text(row.Strength),
		ImageURL:     text(row.ImageUrl),
		Description:  text(row.Description),
		BrandName:    text(row.BrandName),
		CategoryName: text(row.CategoryName),
		UpdatedAt:    row.UpdatedAt,
	}, nil
}

func text(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
