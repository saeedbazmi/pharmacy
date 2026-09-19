package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
)

type stubFetcher struct {
	items []RawItem
	err   error
	calls int
}

func (f *stubFetcher) Fetch(context.Context, Source) ([]RawItem, error) {
	f.calls++
	return f.items, f.err
}

type memoryStore struct {
	sources map[int64]Source
	persist int
}

func (m *memoryStore) Source(_ context.Context, id int64) (Source, error) {
	src, ok := m.sources[id]
	if !ok {
		return Source{}, errors.New("missing source")
	}
	return src, nil
}
func (m *memoryStore) DueSources(context.Context) ([]Source, error) { return nil, nil }
func (m *memoryStore) StartRun(context.Context, int64) (int64, error) {
	return 1, nil
}
func (m *memoryStore) FinishRun(context.Context, int64, string, int, int, string) error {
	return nil
}
func (m *memoryStore) MarkSource(context.Context, int64, string, string) error { return nil }
func (m *memoryStore) PersistItem(context.Context, Source, RawItem) (persistResult, error) {
	m.persist++
	return persistResult{created: true}, nil
}

func TestFailedSourceDoesNotTouchTheOtherFetcher(t *testing.T) {
	good := &stubFetcher{items: []RawItem{{ExternalID: "1", NameFa: "کرم"}}}
	bad := &stubFetcher{err: errors.New("rosha down")}
	reg := NewRegistry()
	reg.Register("darukade", good)
	reg.Register("rosha", bad)

	sources := map[int64]Source{
		1: {ID: 1, Config: json.RawMessage(`{"fetcher":"darukade"}`)},
		2: {ID: 2, Config: json.RawMessage(`{"fetcher":"rosha"}`)},
	}
	store := &memoryStore{sources: sources}
	svc := NewService(store, reg, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	if err := svc.SyncSource(context.Background(), 2); err == nil {
		t.Fatal("rosha sync should fail")
	}
	if good.calls != 0 {
		t.Fatalf("darukade fetcher called %d times during rosha failure", good.calls)
	}
	if err := svc.SyncSource(context.Background(), 1); err != nil {
		t.Fatalf("darukade sync: %v", err)
	}
	if good.calls != 1 || bad.calls != 1 {
		t.Fatalf("calls good=%d bad=%d", good.calls, bad.calls)
	}
	if store.persist != 1 {
		t.Fatalf("persist = %d, want only the healthy source", store.persist)
	}
}
