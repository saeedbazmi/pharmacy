package redirect

import (
	"context"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
)

type offerLookup interface {
	ActiveOffer(ctx context.Context, id int64) (Target, error)
}

// Service resolves a public offer to its stored pharmacy URL.
type Service struct {
	offers   offerLookup
	recorder *Recorder
	log      *slog.Logger
}

func NewService(offers offerLookup, recorder *Recorder, log *slog.Logger) *Service {
	return &Service{offers: offers, recorder: recorder, log: log}
}

// GoTarget returns the stored product URL. Query parameters never influence it.
func (s *Service) GoTarget(ctx context.Context, rawID, referrer string) (Target, error) {
	id, err := parseOfferID(rawID)
	if err != nil {
		return Target{}, err
	}
	target, err := s.offers.ActiveOffer(ctx, id)
	if err != nil {
		return Target{}, err
	}
	if !isHTTPURL(target.ProductURL) {
		return Target{}, ErrOfferNotFound
	}
	s.log.InfoContext(ctx, "redirect.click",
		"offer_id", target.OfferID,
		"product_id", target.ProductID,
		"pharmacy_id", target.PharmacyID)
	s.recorder.Submit(Click{
		OfferID:    target.OfferID,
		ProductID:  target.ProductID,
		PharmacyID: target.PharmacyID,
		Referrer:   referrer,
	})
	return target, nil
}

func parseOfferID(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, ErrInvalidOfferID
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		return 0, ErrInvalidOfferID
	}
	return id, nil
}

func isHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}
