package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/crawlhttp"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/jobq"
)

// Runner claims jobs and executes them with a concurrency cap.
type Runner struct {
	queue   *jobq.Queue
	ingest  *Service
	log     *slog.Logger
	workers int
	extra   map[string]func(context.Context, jobq.Job) error
}

func NewRunner(pool *pgxpool.Pool, ingest *Service, log *slog.Logger, workers int) *Runner {
	if workers < 1 {
		workers = 2
	}
	return &Runner{
		queue:   jobq.New(pool),
		ingest:  ingest,
		log:     log,
		workers: workers,
		extra:   map[string]func(context.Context, jobq.Job) error{},
	}
}

// Handle registers a job kind that is not a source sync.
func (r *Runner) Handle(kind string, fn func(context.Context, jobq.Job) error) {
	r.extra[kind] = fn
}

// ScheduleDue enqueues a sync job for every source whose schedule is due,
// skipping sources that already have an open job.
func (r *Runner) ScheduleDue(ctx context.Context) error {
	sources, err := r.ingest.DueSources(ctx)
	if err != nil {
		return err
	}
	for _, src := range sources {
		ok, err := r.queue.EnqueueSyncSource(ctx, src.ID)
		if err != nil {
			return err
		}
		if ok {
			r.log.InfoContext(ctx, "job.enqueued", "kind", jobq.KindSyncSource, "source_id", src.ID)
		}
	}
	return nil
}

// Run claims and executes jobs until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(r.workers)
	for {
		if ctx.Err() != nil {
			return g.Wait()
		}
		job, err := r.queue.Claim(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				select {
				case <-ctx.Done():
					return g.Wait()
				case <-time.After(2 * time.Second):
				}
				continue
			}
			return fmt.Errorf("claim job: %w", err)
		}
		j := *job
		g.Go(func() error {
			r.execute(ctx, j)
			return nil
		})
	}
}

func (r *Runner) execute(ctx context.Context, job jobq.Job) {
	var runErr error
	switch job.Kind {
	case jobq.KindSyncSource:
		sourceID, err := job.SourceID()
		if err != nil {
			runErr = err
			break
		}
		runErr = r.ingest.SyncSource(ctx, sourceID)
	default:
		if h, ok := r.extra[job.Kind]; ok {
			runErr = h(ctx, job)
			break
		}
		runErr = fmt.Errorf("unknown job kind %q", job.Kind)
	}

	if runErr != nil {
		if crawlhttp.IsStructural(runErr) {
			job.Attempts = job.MaxAttempts
			r.log.ErrorContext(ctx, "job.failed", "job_id", job.ID, "kind", job.Kind, "error", runErr.Error(), "class", "structural")
		} else {
			r.log.WarnContext(ctx, "job.failed", "job_id", job.ID, "kind", job.Kind, "error", runErr.Error(), "class", "transient")
		}
		if err := r.queue.Fail(ctx, job, runErr); err != nil {
			r.log.ErrorContext(ctx, "job.fail_persist", "job_id", job.ID, "error", err.Error())
		}
		return
	}
	if err := r.queue.Complete(ctx, job.ID); err != nil {
		r.log.ErrorContext(ctx, "job.complete_persist", "job_id", job.ID, "error", err.Error())
	}
}
