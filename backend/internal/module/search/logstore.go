package search

import (
	"context"
	"fmt"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

type logStore struct {
	q *store.Queries
}

func NewLogStore(db store.DBTX) *logStore {
	return &logStore{q: store.New(db)}
}

func (s *logStore) InsertQuery(ctx context.Context, e Event) error {
	if err := s.q.InsertSearchQuery(ctx, store.InsertSearchQueryParams{
		QueryNormalized: e.Query,
		ResultCount:     int32(e.ResultCount),
		DurationMs:      int32(e.Duration.Milliseconds()),
		UserID:          e.UserID,
	}); err != nil {
		return fmt.Errorf("insert search query: %w", err)
	}
	return nil
}
