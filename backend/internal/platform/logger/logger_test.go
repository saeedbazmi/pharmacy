package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/reqctx"
)

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"iranian mobile", "09121234567", "0912***4567"},
		{"with separators", "0912-123-4567", "0912***4567"},
		{"with country code", "+989121234567", "9891***4567"},
		{"persian digits are not digits", "۰۹۱۲۱۲۳۴۵۶۷", "***"},
		{"too short", "12345", "***"},
		{"empty", "", "***"},
		{"exactly seven digits", "1234567", "1234***4567"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MaskPhone(tc.input); got != tc.want {
				t.Fatalf("MaskPhone(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestMaskPhoneNeverLeaksMiddleDigits(t *testing.T) {
	const phone = "09121234567"
	masked := MaskPhone(phone)
	if strings.Contains(masked, "123456") {
		t.Fatalf("masked value %q still contains subscriber digits", masked)
	}
}

func TestLoggerEmitsBaseFields(t *testing.T) {
	var buf bytes.Buffer
	log := NewWithWriter(&buf, "api", slog.LevelInfo)

	ctx := reqctx.WithRequestID(context.Background(), "req-123")
	log.InfoContext(ctx, "offer.updated", "product_id", int64(7))

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}

	for _, field := range []string{"time", "level", "msg", "service", "request_id"} {
		if _, ok := record[field]; !ok {
			t.Errorf("log record is missing base field %q", field)
		}
	}
	if record["msg"] != "offer.updated" {
		t.Errorf("msg = %v, want offer.updated", record["msg"])
	}
	if record["service"] != "api" {
		t.Errorf("service = %v, want api", record["service"])
	}
	if record["request_id"] != "req-123" {
		t.Errorf("request_id = %v, want req-123", record["request_id"])
	}
}

func TestLoggerKeepsContextFieldsAfterWith(t *testing.T) {
	var buf bytes.Buffer
	log := NewWithWriter(&buf, "worker", slog.LevelInfo).With("source_id", int64(42))

	ctx := reqctx.WithRequestID(context.Background(), "req-456")
	log.InfoContext(ctx, "sync.completed")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("log line is not valid JSON: %v", err)
	}
	if record["request_id"] != "req-456" {
		t.Fatalf("request_id lost after With(): %v", record)
	}
}

func TestParseLevel(t *testing.T) {
	tests := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"":      slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
	for input, want := range tests {
		got, err := ParseLevel(input)
		if err != nil {
			t.Fatalf("ParseLevel(%q) returned error: %v", input, err)
		}
		if got != want {
			t.Errorf("ParseLevel(%q) = %v, want %v", input, got, want)
		}
	}
	if _, err := ParseLevel("verbose"); err == nil {
		t.Error("ParseLevel(\"verbose\") should fail")
	}
}
