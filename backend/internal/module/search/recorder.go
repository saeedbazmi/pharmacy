package search

import (
	"context"
	"log/slog"
	"time"
)

type eventWriter interface {
	InsertQuery(ctx context.Context, e Event) error
}

type Recorder struct {
	ch   chan Event
	done chan struct{}
	out  eventWriter
	log  *slog.Logger
}

func NewRecorder(out eventWriter, log *slog.Logger) *Recorder {
	r := &Recorder{
		ch:   make(chan Event, 256),
		done: make(chan struct{}),
		out:  out,
		log:  log,
	}
	go r.loop()
	return r
}

func (r *Recorder) Submit(e Event) {
	if r == nil {
		return
	}
	select {
	case r.ch <- e:
	default:
		r.log.Warn("search.query_dropped", "q", e.Query)
	}
}

func (r *Recorder) Stop() {
	if r == nil {
		return
	}
	close(r.ch)
	<-r.done
}

func (r *Recorder) loop() {
	defer close(r.done)
	for e := range r.ch {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := r.out.InsertQuery(ctx, e); err != nil {
			r.log.Warn("search.query_persist", "error", err.Error())
		}
		cancel()
	}
}
