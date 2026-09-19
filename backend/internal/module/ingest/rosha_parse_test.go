package ingest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseRoshaListing(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", "rosha-listing.html"))
	if err != nil {
		t.Fatal(err)
	}
	items, err := ParseRoshaListing(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if items[0].ExternalID != "4197" || items[0].NameFa != "کرم دور چشم دکتر ژیلا" {
		t.Fatalf("first = %+v", items[0])
	}
	if items[0].PriceRial() != 6136000 {
		t.Errorf("price rial = %d", items[0].PriceRial())
	}
	if !items[0].InStock {
		t.Error("first should be in stock")
	}
	if items[0].ProductURL != "https://roshapharmacy.com/product/4197" {
		t.Errorf("url = %s", items[0].ProductURL)
	}
	if items[1].InStock {
		t.Error("second should be out of stock")
	}
	if items[1].ExternalID != "999" {
		t.Errorf("second id = %s", items[1].ExternalID)
	}
	if !json.Valid(items[0].Raw) {
		t.Fatalf("raw is not JSON: %s", items[0].Raw)
	}
}

func TestParseRoshaListingRejectsEmpty(t *testing.T) {
	_, err := ParseRoshaListing([]byte(`<html><body>nothing</body></html>`))
	if err == nil {
		t.Fatal("empty listing should fail")
	}
}
