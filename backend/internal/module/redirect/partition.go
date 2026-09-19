package redirect

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

// EnsureMonths creates this month's and next month's click partitions.
func EnsureMonths(ctx context.Context, q *store.Queries) error {
	now := time.Now().UTC()
	this := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	next := this.AddDate(0, 1, 0)
	if err := q.EnsureRedirectClicksMonth(ctx, pgtype.Date{Time: this, Valid: true}); err != nil {
		return fmt.Errorf("ensure click partition %s: %w", this.Format("2006-01"), err)
	}
	if err := q.EnsureRedirectClicksMonth(ctx, pgtype.Date{Time: next, Valid: true}); err != nil {
		return fmt.Errorf("ensure click partition %s: %w", next.Format("2006-01"), err)
	}
	return nil
}
