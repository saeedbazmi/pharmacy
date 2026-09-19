package search

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/cache"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

type fakePort struct {
	res   Result
	err   error
	got   Query
	calls int
}

func (f *fakePort) Search(_ context.Context, q Query) (Result, error) {
	f.calls++
	f.got = q
	return f.res, f.err
}

func (f *fakePort) Facets(context.Context) ([]Facet, []Facet, error) {
	return []Facet{{Slug: "dr-jila", Name: "Dr Jila"}}, nil, nil
}

func TestParseQueryUsesTextfaAndDefaults(t *testing.T) {
	q, err := parseQuery(url.Values{
		"q":        {"استامينوفن"},
		"sort":     {"nope"},
		"in_stock": {"maybe"},
		"page":     {"-3"},
		"brand":    {"../x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if q.Q != textfa.Query("استامینوفن") {
		t.Fatalf("q = %q", q.Q)
	}
	if q.Sort != SortRelevance || q.InStock || q.Brand != "" || q.Page != 1 || q.PageSize != 24 {
		t.Fatalf("%+v", q)
	}
}

func TestParseQueryRejectsEmpty(t *testing.T) {
	if _, err := parseQuery(url.Values{"q": {"   "}}); err != ErrQueryEmpty {
		t.Fatalf("err = %v", err)
	}
}

func TestRankBucketExactThenPrefix(t *testing.T) {
	if rankBucket("استامینوفن 500", "استامینوفن 500", "استامینوفن 500 acetaminophen") != 0 {
		t.Fatal("exact")
	}
	if rankBucket("استامینوفن", "استامینوفن 500", "استامینوفن 500") != 1 {
		t.Fatal("prefix")
	}
}

func TestSearchCachesAndLogs(t *testing.T) {
	port := &fakePort{res: Result{Hits: []Hit{{ID: 1, Slug: "acetaminophen-500", NameFa: "استامینوفن ۵۰۰"}}, Total: 1}}
	svc := NewService(port, nil, cache.NewTTL[Result](time.Minute, 8), slog.New(slog.NewJSONHandler(io.Discard, nil)))
	q := Query{Q: "استامینوفن", Sort: SortRelevance, Page: 1, PageSize: 24}
	first, err := svc.Search(context.Background(), q)
	if err != nil || first.Total != 1 {
		t.Fatalf("%+v %v", first, err)
	}
	if _, err := svc.Search(context.Background(), q); err != nil {
		t.Fatal(err)
	}
	if port.calls != 1 {
		t.Fatalf("port calls = %d, cache should skip the second", port.calls)
	}
}
