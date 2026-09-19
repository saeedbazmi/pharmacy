package alerting

import "testing"

func TestShouldFirePriceDrop(t *testing.T) {
	target := int64(80000)
	if !ShouldFirePriceDrop(100000, nil, 90000) {
		t.Fatal("drop below baseline should fire")
	}
	if ShouldFirePriceDrop(100000, nil, 100000) {
		t.Fatal("same price is not a drop")
	}
	if ShouldFirePriceDrop(100000, nil, 0) {
		t.Fatal("missing in-stock quote must not fire")
	}
	if ShouldFirePriceDrop(100000, &target, 90000) {
		t.Fatal("drop that is still above the target should wait")
	}
	if !ShouldFirePriceDrop(100000, &target, 80000) {
		t.Fatal("reaching the target should fire")
	}
}

func TestShouldFireBackInStock(t *testing.T) {
	if !ShouldFireBackInStock(false, true) {
		t.Fatal("out → in should fire")
	}
	if ShouldFireBackInStock(true, true) {
		t.Fatal("still in stock should not fire")
	}
	if ShouldFireBackInStock(false, false) {
		t.Fatal("still out should not fire")
	}
}

func TestDropEventKeyIncludesPrice(t *testing.T) {
	if DropEventKey(12000) != "drop:12000" {
		t.Fatal(DropEventKey(12000))
	}
}
