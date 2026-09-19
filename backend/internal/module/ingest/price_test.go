package ingest

import "testing"

func TestIsSuspiciousJump(t *testing.T) {
	const threshold = 0.70
	tests := []struct {
		name string
		old  int64
		new  int64
		want bool
	}{
		{"first price", 0, 120_000, false},
		{"first price negative old", -1, 50_000, false},
		{"zero to zero", 0, 0, false},
		{"unchanged", 100_000, 100_000, false},
		{"small rise 10%", 100_000, 110_000, false},
		{"small drop 20%", 100_000, 80_000, false},
		{"just under threshold", 100_000, 170_000, false},
		{"just over threshold", 100_000, 170_001, true},
		{"exact 70 percent not a jump", 100_000, 170_000, false},
		{"double price", 50_000, 100_000, true},
		{"drop to a quarter", 100_000, 25_000, true},
		{"drop to zero", 80_000, 0, true},
		{"tiny change 1 rial", 100_000, 100_001, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsSuspiciousJump(tc.old, tc.new, threshold)
			if got != tc.want {
				t.Fatalf("IsSuspiciousJump(%d, %d, %g) = %v, want %v", tc.old, tc.new, threshold, got, tc.want)
			}
		})
	}
}

func TestIsSuspiciousJumpDisabledThreshold(t *testing.T) {
	if IsSuspiciousJump(100, 1_000_000, 0) {
		t.Fatal("threshold <= 0 should never mark a jump")
	}
}
