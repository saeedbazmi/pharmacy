package search

import (
	"context"
	"log/slog"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/cache"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

type Service struct {
	port  Port
	rec   *Recorder
	cache *cache.TTL[Result]
	log   *slog.Logger
}

func NewService(port Port, rec *Recorder, mem *cache.TTL[Result], log *slog.Logger) *Service {
	return &Service{port: port, rec: rec, cache: mem, log: log}
}

func (s *Service) Search(ctx context.Context, q Query) (Result, error) {
	start := time.Now()
	if hit, ok := s.cache.Get(cacheKey(q)); ok {
		s.finish(ctx, q, hit, start)
		return hit, nil
	}
	res, err := s.port.Search(ctx, q)
	if err != nil {
		return Result{}, err
	}
	brands, cats, err := s.port.Facets(ctx)
	if err != nil {
		return Result{}, err
	}
	res.Brands = brands
	res.Categories = cats
	s.cache.Set(cacheKey(q), res)
	s.finish(ctx, q, res, start)
	return res, nil
}

func (s *Service) finish(ctx context.Context, q Query, res Result, start time.Time) {
	dur := time.Since(start)
	s.log.InfoContext(ctx, "search.query",
		"q", q.Q,
		"result_count", res.Total,
		"duration_ms", dur.Milliseconds(),
		"page", q.Page)
	if s.rec != nil {
		var uid *int64
		if u, ok := reqctx.UserFrom(ctx); ok {
			id := u.ID
			uid = &id
		}
		s.rec.Submit(Event{Query: q.Q, ResultCount: res.Total, Duration: dur, UserID: uid})
	}
}
