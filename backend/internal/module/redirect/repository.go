package redirect

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

type Repository struct {
	q *store.Queries
}

func NewRepository(db store.DBTX) *Repository {
	return &Repository{q: store.New(db)}
}

func (r *Repository) ActiveOffer(ctx context.Context, id int64) (Target, error) {
	row, err := r.q.GetActiveOfferByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Target{}, ErrOfferNotFound
		}
		return Target{}, fmt.Errorf("get offer %d: %w", id, err)
	}
	return Target{
		OfferID:    row.ID,
		ProductID:  row.ProductID,
		PharmacyID: row.PharmacyID,
		ProductURL: row.ProductUrl,
	}, nil
}

func (r *Repository) InsertClick(ctx context.Context, c Click) error {
	if err := r.q.InsertRedirectClick(ctx, store.InsertRedirectClickParams{
		OfferID:    c.OfferID,
		ProductID:  c.ProductID,
		PharmacyID: c.PharmacyID,
		Referrer:   strPtr(c.Referrer),
	}); err != nil {
		return fmt.Errorf("insert redirect click: %w", err)
	}
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
