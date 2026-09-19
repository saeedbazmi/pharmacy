package redirect

import (
	"context"
	"log/slog"
	"time"
)

type clickWriter interface {
	InsertClick(ctx context.Context, c Click) error
}

// Recorder accepts clicks without blocking the redirect response.
type Recorder struct {
	ch   chan Click
	done chan struct{}
	out  clickWriter
	log  *slog.Logger
}

func NewRecorder(out clickWriter, log *slog.Logger) *Recorder {
	r := &Recorder{
		ch:   make(chan Click, 256),
		done: make(chan struct{}),
		out:  out,
		log:  log,
	}
	go r.loop()
	return r
}

func (r *Recorder) Submit(c Click) {
	if r == nil {
		return
	}
	select {
	case r.ch <- c:
	default:
		r.log.Warn("redirect.click_dropped", "offer_id", c.OfferID)
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
	for c := range r.ch {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := r.out.InsertClick(ctx, c); err != nil {
			r.log.Warn("redirect.click_persist", "error", err.Error(), "offer_id", c.OfferID)
		}
		cancel()
	}
}
