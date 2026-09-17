// Package httpx holds the small transport helpers shared by handlers: JSON
// encoding, bounded request decoding and status capture.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// WriteJSON writes v as a JSON body with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// The status line is already sent, so the only thing left is to stop
		// writing. The access log records the response either way.
		return
	}
}

// ErrBodyTooLarge is returned when a request body exceeds the configured limit.
var ErrBodyTooLarge = errors.New("request body too large")

// ReadJSON decodes a bounded request body into dst and rejects unknown fields
// so a typo in a client payload fails loudly instead of being ignored.
func ReadJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return ErrBodyTooLarge
		}
		return fmt.Errorf("decode json body: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("body must contain a single json object")
	}
	return nil
}

// StatusRecorder captures the status code and size for the access log.
type StatusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

// NewStatusRecorder wraps w so the written status can be inspected afterwards.
func NewStatusRecorder(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{ResponseWriter: w}
}

// WriteHeader records the status before delegating.
func (r *StatusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

// Write records the body size, defaulting the status to 200 like net/http does.
func (r *StatusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Status returns the written status code, or 200 when nothing was written.
func (r *StatusRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

// Bytes returns the number of body bytes written.
func (r *StatusRecorder) Bytes() int { return r.bytes }

// Unwrap lets http.ResponseController reach the underlying writer.
func (r *StatusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
