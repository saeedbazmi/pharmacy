package ingest

import "testing"

func TestDecideMatch(t *testing.T) {
	item := RawItem{NameFa: "کرم دور چشم", GTIN: "", IRC: ""}
	linked := &ProductRef{ID: 9, Slug: "x"}

	t.Run("gtin wins", func(t *testing.T) {
		got := DecideMatch(RawItem{GTIN: "123", NameFa: "a"}, linked, nil, nil)
		if got.Action != ActionLink || got.Reason != "gtin" || got.ProductID != 9 {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("irc wins", func(t *testing.T) {
		got := DecideMatch(RawItem{IRC: "IRC1"}, nil, linked, nil)
		if got.Action != ActionLink || got.Reason != "irc" {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("exact name", func(t *testing.T) {
		got := DecideMatch(item, nil, nil, linked)
		if got.Action != ActionLink || got.Reason != "exact_name" {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("first seen creates", func(t *testing.T) {
		got := DecideMatch(item, nil, nil, nil)
		if got.Action != ActionCreate || got.Reason != "first_seen" {
			t.Fatalf("%+v", got)
		}
	})
	t.Run("same name different dose stays exact-only", func(t *testing.T) {
		// Matching is exact on the already-normalised name looked up by the
		// caller. A 500mg vs 250mg item would have a different NameFa, so
		// byName is nil and we create a new product instead of merging.
		got := DecideMatch(RawItem{NameFa: "استامینوفن 500"}, nil, nil, nil)
		if got.Action != ActionCreate {
			t.Fatalf("different dose must not merge: %+v", got)
		}
	})
	t.Run("code present but unknown is queued", func(t *testing.T) {
		got := DecideMatch(RawItem{GTIN: "999", NameFa: "x"}, nil, nil, nil)
		if got.Action != ActionQueue {
			t.Fatalf("%+v", got)
		}
	})
}
