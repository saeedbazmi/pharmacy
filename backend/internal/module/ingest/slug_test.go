package ingest

import "testing"

func TestProductSlugStableWithID(t *testing.T) {
	item := RawItem{ExternalID: "1664", SlugHint: "cream-dr-jila-eye-contour-cream", NameFa: "نام جدید"}
	got := productSlug(item)
	if got != "cream-dr-jila-eye-contour-cream-1664" {
		t.Fatalf("got %q", got)
	}
	item.NameFa = "یک اسم دیگر"
	if productSlug(item) != got {
		t.Fatal("changing the Persian name must not change the slug")
	}
}

func TestUniqueSlugSuffix(t *testing.T) {
	taken := map[string]bool{"a-1": true, "a-1-2": true}
	got := uniqueSlug("a-1", func(s string) bool { return taken[s] })
	if got != "a-1-3" {
		t.Fatalf("got %q", got)
	}
}
