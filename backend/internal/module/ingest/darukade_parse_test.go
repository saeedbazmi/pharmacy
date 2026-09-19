package ingest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdata(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestParseListingJSON(t *testing.T) {
	items, err := ParseListing(testdata(t, "listing-sample.json"), "https://darukade.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	first := items[0]
	if first.ExternalID != "27774" {
		t.Errorf("external id = %s", first.ExternalID)
	}
	if first.PriceRial() != 3464000 {
		t.Errorf("price rial = %d, want toman*10", first.PriceRial())
	}
	if first.InStock {
		t.Error("first product should be out of stock")
	}
	if first.ProductURL != "https://darukade.com/products/cosmetic-eye-and-lip-anti-wrinkle-763/cream-wrinkle-repair-and-lift-eye-cream-27774" {
		t.Errorf("url = %s", first.ProductURL)
	}
	if first.SlugHint != "cream-wrinkle-repair-and-lift-eye-cream" {
		t.Errorf("slug hint = %s", first.SlugHint)
	}

	second := items[1]
	if second.PriceRial() != 6478000 {
		t.Errorf("discounted rial = %d", second.PriceRial())
	}
	if !second.InStock {
		t.Error("second product should be in stock")
	}
	if second.BrandName != "Dr Jila" {
		t.Errorf("brand = %s", second.BrandName)
	}
}

func TestParseListingHTML(t *testing.T) {
	items, err := ParseListing(testdata(t, "listing-sample.html"), "https://darukade.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ExternalID != "27774" {
		t.Fatalf("html parse = %+v", items)
	}
}

func TestParseListingRejectsEmpty(t *testing.T) {
	_, err := ParseListing([]byte(`{"props":{"pageProps":{"categoryMataData":{"products":[]}}}}`), "https://darukade.com")
	if err == nil {
		t.Fatal("empty listing should fail")
	}
}
