package identity

import (
	"testing"
)

func TestNormalizePhoneAcceptsCommonWritings(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"09121234567", "09121234567"},
		{"9121234567", "09121234567"},
		{"989121234567", "09121234567"},
		{"00989121234567", "09121234567"},
		{"۰۹۱۲۱۲۳۴۵۶۷", "09121234567"},
		{"0912 123 4567", "09121234567"},
	}
	for _, tc := range tests {
		got, err := NormalizePhone(tc.in)
		if err != nil {
			t.Fatalf("NormalizePhone(%q) error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("NormalizePhone(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizePhoneRejectsInvalid(t *testing.T) {
	for _, in := range []string{"", "123", "02122334455", "0912123456", "08121234567"} {
		if _, err := NormalizePhone(in); err != ErrInvalidPhone {
			t.Fatalf("NormalizePhone(%q) = %v, want ErrInvalidPhone", in, err)
		}
	}
}

func TestMaskPhoneHidesSubscriber(t *testing.T) {
	got := MaskPhone("09121234567")
	if got != "0912***4567" {
		t.Fatalf("MaskPhone = %q", got)
	}
	if got == "09121234567" {
		t.Fatal("full number leaked")
	}
}
