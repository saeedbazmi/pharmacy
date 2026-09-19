package ops

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMaskSecretsHidesAPIKeys(t *testing.T) {
	raw := json.RawMessage(`{"fetcher":"darukade","listing_url":"https://darukade.com/x","api_key":"super-secret"}`)
	got := MaskSecrets(raw)
	if strings.Contains(string(got), "super-secret") {
		t.Fatalf("secret leaked: %s", got)
	}
	if !strings.Contains(string(got), maskedSecret) {
		t.Fatalf("expected mask, got %s", got)
	}
}

func TestMergeSecretsKeepsStoredKeyWhenMasked(t *testing.T) {
	stored := json.RawMessage(`{"fetcher":"darukade","listing_url":"https://darukade.com/x","api_key":"real-key"}`)
	incoming := json.RawMessage(`{"fetcher":"darukade","listing_url":"https://darukade.com/x","api_key":"********"}`)
	got := MergeSecrets(stored, incoming)
	if !strings.Contains(string(got), "real-key") {
		t.Fatalf("stored key was dropped: %s", got)
	}
}

func TestParseSourceConfigRejectsPrivateURL(t *testing.T) {
	_, err := parseSourceConfig(json.RawMessage(`{"fetcher":"darukade","listing_url":"http://127.0.0.1/secret"}`), []string{"darukade"})
	if err == nil {
		t.Fatal("expected SSRF rejection")
	}
}

func TestParseSourceConfigAcceptsPublicHTTPS(t *testing.T) {
	cfg, err := parseSourceConfig(json.RawMessage(`{"fetcher":"darukade","listing_url":"https://darukade.com/products"}`), []string{"darukade"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxPages != 1 {
		t.Fatalf("max_pages = %d", cfg.MaxPages)
	}
}
