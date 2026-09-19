package ops

import (
	"bytes"
	"encoding/json"
	"strings"
)

var secretKeys = map[string]struct{}{
	"api_key":       {},
	"apikey":        {},
	"api-key":       {},
	"token":         {},
	"access_token":  {},
	"password":      {},
	"secret":        {},
	"authorization": {},
	"auth":          {},
}

const maskedSecret = "********"

// MaskSecrets walks a JSON document and replaces secret field values.
// Used for panel reads and audit payloads so an API key is never returned.
func MaskSecrets(raw json.RawMessage) json.RawMessage {
	if len(bytes.TrimSpace(raw)) == 0 {
		return json.RawMessage(`{}`)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return json.RawMessage(`{}`)
	}
	masked := maskValue(v)
	out, err := json.Marshal(masked)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return out
}

// MergeSecrets copies secret values from stored into incoming when the
// incoming value is empty or already masked, so an edit does not wipe a key.
func MergeSecrets(stored, incoming json.RawMessage) json.RawMessage {
	var src, dst map[string]any
	if err := json.Unmarshal(nonzeroJSON(stored), &src); err != nil {
		src = map[string]any{}
	}
	if err := json.Unmarshal(nonzeroJSON(incoming), &dst); err != nil {
		return MaskSecrets(stored)
	}
	for key, val := range src {
		if !isSecretKey(key) {
			continue
		}
		incomingVal, ok := dst[key]
		if !ok || isMaskedOrEmpty(incomingVal) {
			dst[key] = val
		}
	}
	out, err := json.Marshal(dst)
	if err != nil {
		return incoming
	}
	return out
}

func maskValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, child := range t {
			if isSecretKey(k) {
				if s, ok := child.(string); ok && s != "" {
					out[k] = maskedSecret
					continue
				}
			}
			out[k] = maskValue(child)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, child := range t {
			out[i] = maskValue(child)
		}
		return out
	default:
		return v
	}
}

func isSecretKey(key string) bool {
	_, ok := secretKeys[strings.ToLower(strings.TrimSpace(key))]
	return ok
}

func isMaskedOrEmpty(v any) bool {
	s, ok := v.(string)
	if !ok {
		return v == nil
	}
	s = strings.TrimSpace(s)
	return s == "" || s == maskedSecret
}

func nonzeroJSON(raw json.RawMessage) []byte {
	if len(bytes.TrimSpace(raw)) == 0 {
		return []byte(`{}`)
	}
	return raw
}
