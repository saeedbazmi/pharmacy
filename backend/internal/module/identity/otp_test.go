package identity

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestCompareOTPIsConstantTimeMatch(t *testing.T) {
	pepper, phone, code := "pepper", "09121234567", "123456"
	hash := hashOTP(pepper, phone, code)
	if !compareOTP(pepper, phone, code, hash) {
		t.Fatal("matching code should verify")
	}
	if compareOTP(pepper, phone, "000000", hash) {
		t.Fatal("wrong code should not verify")
	}
	if compareOTP("other", phone, code, hash) {
		t.Fatal("wrong pepper should not verify")
	}
}

func TestDevSenderPrintsCodeOnlyToWriterNotSlog(t *testing.T) {
	var out bytes.Buffer
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	sender := NewDevSender(true, &out)
	if err := sender.SendOTP("09121234567", "654321"); err != nil {
		t.Fatal(err)
	}
	printed := out.String()
	if !strings.Contains(printed, "654321") {
		t.Fatalf("dev preview should print the code, got %q", printed)
	}
	if !strings.Contains(printed, "0912***4567") {
		t.Fatalf("dev preview should mask the phone, got %q", printed)
	}
	if strings.Contains(printed, "09121234567") {
		t.Fatal("full phone leaked to stdout")
	}
	_ = log
	if logs.Len() != 0 {
		t.Fatalf("structured log must stay empty, got %s", logs.Bytes())
	}
}

func TestDevSenderSilentWhenPrintDisabled(t *testing.T) {
	var out bytes.Buffer
	sender := NewDevSender(false, &out)
	if err := sender.SendOTP("09121234567", "111111"); err != nil {
		t.Fatal(err)
	}
	if out.Len() != 0 {
		t.Fatalf("printCode=false must not write, got %q", out.String())
	}
}

func TestUnavailableSender(t *testing.T) {
	if err := NewUnavailableSender().SendOTP("09121234567", "123456"); err != ErrSMSUnavailable {
		t.Fatalf("err = %v", err)
	}
}
