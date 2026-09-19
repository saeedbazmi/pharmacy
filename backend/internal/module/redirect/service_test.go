package redirect

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeOffers struct {
	target Target
	err    error
	gotID  int64
}

func (f *fakeOffers) ActiveOffer(_ context.Context, id int64) (Target, error) {
	f.gotID = id
	return f.target, f.err
}

type fakeClicks struct{ n int }

func (f *fakeClicks) InsertClick(context.Context, Click) error {
	f.n++
	return nil
}

func TestGoTargetIgnoresQueryURL(t *testing.T) {
	offers := &fakeOffers{target: Target{
		OfferID:    9,
		ProductID:  3,
		PharmacyID: 1,
		ProductURL: "https://roshapharmacy.com/product/4197",
	}}
	clicks := &fakeClicks{}
	rec := NewRecorder(clicks, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	defer rec.Stop()
	svc := NewService(offers, rec, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	got, err := svc.GoTarget(context.Background(), "9", "https://evil.example/?url=https://evil.example")
	if err != nil {
		t.Fatal(err)
	}
	if got.ProductURL != "https://roshapharmacy.com/product/4197" {
		t.Fatalf("redirected to %s", got.ProductURL)
	}
	if offers.gotID != 9 {
		t.Fatalf("looked up %d", offers.gotID)
	}
}

func TestGoTargetRejectsNonHTTPStoredURL(t *testing.T) {
	offers := &fakeOffers{target: Target{OfferID: 1, ProductURL: "javascript:alert(1)"}}
	svc := NewService(offers, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	_, err := svc.GoTarget(context.Background(), "1", "")
	if !errors.Is(err, ErrOfferNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestGoTargetInvalidID(t *testing.T) {
	svc := NewService(&fakeOffers{}, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	_, err := svc.GoTarget(context.Background(), "nope", "")
	if !errors.Is(err, ErrInvalidOfferID) {
		t.Fatalf("err = %v", err)
	}
}

func TestGoHandlerRedirectsToStoredURL(t *testing.T) {
	offers := &fakeOffers{target: Target{
		OfferID:    9,
		ProductURL: "https://darukade.com/products/x",
	}}
	svc := NewService(offers, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	h := NewHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/go/9?url=https://evil.example", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "https://darukade.com/products/x" {
		t.Fatalf("location = %s", loc)
	}
}

func TestGoHandlerUnknownOfferIsPersian(t *testing.T) {
	offers := &fakeOffers{err: ErrOfferNotFound}
	svc := NewService(offers, nil, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	h := NewHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/go/99", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("این پیشنهاد دیگر معتبر نیست")) {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
