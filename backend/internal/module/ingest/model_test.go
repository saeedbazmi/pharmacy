package ingest

import "testing"

func TestPriceRialConvertsToman(t *testing.T) {
	item := RawItem{PriceToman: 346400}
	if got := item.PriceRial(); got != 3464000 {
		t.Fatalf("PriceRial() = %d, want 3464000", got)
	}
}

func TestPriceRialRejectsNegative(t *testing.T) {
	item := RawItem{PriceToman: -1}
	if got := item.PriceRial(); got != 0 {
		t.Fatalf("negative toman stored as %d, want 0", got)
	}
}
