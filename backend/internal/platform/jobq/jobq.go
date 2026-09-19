// Package jobq is the Postgres-backed work queue. Claim uses
// FOR UPDATE SKIP LOCKED so two workers cannot take the same row.
package jobq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

const (
	KindSyncSource       = "sync_source"
	KindEnsurePartitions = "ensure_click_partitions"
	KindEvaluateAlerts   = "evaluate_alerts"
	defaultMaxTries      = 5
)

// Job is one unit of background work.
type Job struct {
	ID          int64
	Kind        string
	Payload     json.RawMessage
	Attempts    int32
	MaxAttempts int32
}

// Queue enqueues, claims and settles jobs.
type Queue struct {
	q *store.Queries
}

func New(db store.DBTX) *Queue {
	return &Queue{q: store.New(db)}
}

// EnqueueSyncSource records a sync job unless one is already open for the source.
func (q *Queue) EnqueueSyncSource(ctx context.Context, sourceID int64) (enqueued bool, err error) {
	exists, err := q.q.HasOpenJob(ctx, store.HasOpenJobParams{
		Kind:     KindSyncSource,
		SourceID: fmt.Sprintf("%d", sourceID),
	})
	if err != nil {
		return false, fmt.Errorf("check open job: %w", err)
	}
	if exists {
		return false, nil
	}
	payload, err := json.Marshal(map[string]int64{"source_id": sourceID})
	if err != nil {
		return false, fmt.Errorf("encode payload: %w", err)
	}
	_, err = q.q.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:        KindSyncSource,
		Payload:     payload,
		MaxAttempts: defaultMaxTries,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		return false, fmt.Errorf("enqueue sync_source %d: %w", sourceID, err)
	}
	return true, nil
}

// EnqueueEnsurePartitions records a partition job unless one is already open.
func (q *Queue) EnqueueEnsurePartitions(ctx context.Context) (bool, error) {
	exists, err := q.q.HasOpenJobByKind(ctx, KindEnsurePartitions)
	if err != nil {
		return false, fmt.Errorf("check open partition job: %w", err)
	}
	if exists {
		return false, nil
	}
	_, err = q.q.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:        KindEnsurePartitions,
		Payload:     []byte("{}"),
		MaxAttempts: defaultMaxTries,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		return false, fmt.Errorf("enqueue ensure partitions: %w", err)
	}
	return true, nil
}

// EnqueueEvaluateAlerts records an alert-evaluation job unless one is already open.
func (q *Queue) EnqueueEvaluateAlerts(ctx context.Context) (bool, error) {
	exists, err := q.q.HasOpenJobByKind(ctx, KindEvaluateAlerts)
	if err != nil {
		return false, fmt.Errorf("check open alert job: %w", err)
	}
	if exists {
		return false, nil
	}
	_, err = q.q.EnqueueJob(ctx, store.EnqueueJobParams{
		Kind:        KindEvaluateAlerts,
		Payload:     []byte("{}"),
		MaxAttempts: defaultMaxTries,
		RunAt:       time.Now().UTC(),
	})
	if err != nil {
		return false, fmt.Errorf("enqueue evaluate alerts: %w", err)
	}
	return true, nil
}

// Claim takes the next due job, or nil when the queue is empty.
func (q *Queue) Claim(ctx context.Context) (*Job, error) {
	row, err := q.q.ClaimJob(ctx)
	if err != nil {
		return nil, err
	}
	return &Job{
		ID:          row.ID,
		Kind:        row.Kind,
		Payload:     row.Payload,
		Attempts:    row.Attempts,
		MaxAttempts: row.MaxAttempts,
	}, nil
}

// Complete marks a job succeeded.
func (q *Queue) Complete(ctx context.Context, id int64) error {
	if err := q.q.CompleteJob(ctx, id); err != nil {
		return fmt.Errorf("complete job %d: %w", id, err)
	}
	return nil
}

// Fail records a retry or a final failure. Retry delay is exponential.
func (q *Queue) Fail(ctx context.Context, job Job, cause error) error {
	status := "retry"
	runAt := time.Now().UTC().Add(Backoff(job.Attempts))
	if job.Attempts >= job.MaxAttempts {
		status = "failed"
		runAt = time.Now().UTC()
	}
	if err := q.q.FailJob(ctx, store.FailJobParams{
		Status:    status,
		RunAt:     runAt,
		LastError: ptr(cause.Error()),
		ID:        job.ID,
	}); err != nil {
		return fmt.Errorf("fail job %d: %w", job.ID, err)
	}
	return nil
}

// SourceID extracts source_id from a sync_source payload.
func (j Job) SourceID() (int64, error) {
	var p struct {
		SourceID int64 `json:"source_id"`
	}
	if err := json.Unmarshal(j.Payload, &p); err != nil {
		return 0, fmt.Errorf("decode job payload: %w", err)
	}
	if p.SourceID == 0 {
		return 0, fmt.Errorf("job %d missing source_id", j.ID)
	}
	return p.SourceID, nil
}

// Backoff returns 30s, 60s, 120s, ... capped at 15 minutes.
func Backoff(attempts int32) time.Duration {
	if attempts < 1 {
		attempts = 1
	}
	d := 30 * time.Second * time.Duration(1<<min(uint(attempts-1), 8))
	if d > 15*time.Minute {
		return 15 * time.Minute
	}
	return d
}

func ptr(s string) *string { return &s }
