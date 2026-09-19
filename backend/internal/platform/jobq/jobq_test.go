package jobq

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestBackoffGrowsThenCaps(t *testing.T) {
	if got := Backoff(1); got.Seconds() != 30 {
		t.Fatalf("attempt 1 = %s, want 30s", got)
	}
	if got := Backoff(2); got.Seconds() != 60 {
		t.Fatalf("attempt 2 = %s, want 60s", got)
	}
	if got := Backoff(3); got.Seconds() != 120 {
		t.Fatalf("attempt 3 = %s, want 120s", got)
	}
	if got := Backoff(20); got.Minutes() != 15 {
		t.Fatalf("attempt 20 = %s, want 15m cap", got)
	}
}

func TestJobSourceID(t *testing.T) {
	j := Job{ID: 1, Payload: []byte(`{"source_id": 42}`)}
	id, err := j.SourceID()
	if err != nil {
		t.Fatal(err)
	}
	if id != 42 {
		t.Fatalf("id = %d", id)
	}
}

func TestClaimSQLUsesSkipLocked(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "..", "..", "db", "queries", "jobs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte("FOR UPDATE SKIP LOCKED")) {
		t.Fatal("ClaimJob must lock with FOR UPDATE SKIP LOCKED so two workers cannot take the same row")
	}
}
