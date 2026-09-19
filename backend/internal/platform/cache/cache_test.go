package cache

import (
	"testing"
	"time"
)

func TestTTLStoresAndExpires(t *testing.T) {
	c := NewTTL[string](20*time.Millisecond, 8)
	c.Set("q", "hit")
	if got, ok := c.Get("q"); !ok || got != "hit" {
		t.Fatalf("got %q %v", got, ok)
	}
	time.Sleep(30 * time.Millisecond)
	if _, ok := c.Get("q"); ok {
		t.Fatal("expired key still present")
	}
}

func TestTTLDeleteRemovesKey(t *testing.T) {
	c := NewTTL[string](time.Minute, 8)
	c.Set("q", "hit")
	c.Delete("q")
	if _, ok := c.Get("q"); ok {
		t.Fatal("deleted key still present")
	}
}
