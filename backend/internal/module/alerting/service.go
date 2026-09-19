package alerting

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
)

// Notifier delivers an in-app/SMS message. Implementations must never log the
// full phone number or include it in structured fields unmasked.
type Notifier interface {
	Notify(phone, message string) error
}

type Service struct {
	store      Store
	sms        Notifier
	log        *slog.Logger
	freshAfter time.Duration
	now        func() time.Time
}

func NewService(store Store, sms Notifier, freshAfter time.Duration, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	if freshAfter <= 0 {
		freshAfter = 24 * time.Hour
	}
	return &Service{store: store, sms: sms, log: log, freshAfter: freshAfter, now: time.Now}
}

func (s *Service) AddFavorite(ctx context.Context, userID, productID int64) error {
	ok, err := s.store.ProductPublished(ctx, productID)
	if err != nil {
		return fmt.Errorf("product exists: %w", err)
	}
	if !ok {
		return ErrProductNotFound
	}
	if err := s.store.UpsertFavorite(ctx, userID, productID); err != nil {
		return fmt.Errorf("upsert favorite: %w", err)
	}
	return nil
}

func (s *Service) RemoveFavorite(ctx context.Context, userID, productID int64) error {
	if err := s.store.DeleteFavorite(ctx, userID, productID); err != nil {
		return fmt.Errorf("delete favorite: %w", err)
	}
	return nil
}

func (s *Service) FavoriteExists(ctx context.Context, userID, productID int64) (bool, error) {
	ok, err := s.store.FavoriteExists(ctx, userID, productID)
	if err != nil {
		return false, fmt.Errorf("favorite exists: %w", err)
	}
	return ok, nil
}

func (s *Service) ListFavorites(ctx context.Context, userID int64) ([]Favorite, error) {
	items, err := s.store.ListFavorites(ctx, userID, s.freshSince())
	if err != nil {
		return nil, fmt.Errorf("list favorites: %w", err)
	}
	return items, nil
}

func (s *Service) CreateAlert(ctx context.Context, userID int64, in AlertInput) (Alert, error) {
	if in.Kind != KindPriceDrop && in.Kind != KindBackInStock {
		return Alert{}, ErrInvalidAlertKind
	}
	if in.ProductID <= 0 {
		return Alert{}, ErrInvalidInput
	}
	if in.TargetPriceRial != nil && *in.TargetPriceRial < 0 {
		return Alert{}, ErrInvalidInput
	}
	ok, err := s.store.ProductPublished(ctx, in.ProductID)
	if err != nil {
		return Alert{}, fmt.Errorf("product exists: %w", err)
	}
	if !ok {
		return Alert{}, ErrProductNotFound
	}
	n, err := s.store.CountActiveAlerts(ctx, userID)
	if err != nil {
		return Alert{}, fmt.Errorf("count alerts: %w", err)
	}
	if n >= MaxActiveAlerts {
		return Alert{}, ErrAlertLimit
	}
	quotes, err := s.store.ListProductQuotes(ctx, []int64{in.ProductID}, s.freshSince())
	if err != nil {
		return Alert{}, fmt.Errorf("quote: %w", err)
	}
	var baseline int64
	stock := false
	if len(quotes) > 0 {
		baseline = quotes[0].LowestInStockRial
		stock = quotes[0].InStock
	}
	alert, err := s.store.InsertAlert(ctx, userID, in, baseline, stock)
	if err != nil {
		return Alert{}, fmt.Errorf("insert alert: %w", err)
	}
	return alert, nil
}

func (s *Service) ListAlerts(ctx context.Context, userID int64) ([]Alert, error) {
	items, err := s.store.ListUserAlerts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}
	return items, nil
}

func (s *Service) DeleteAlert(ctx context.Context, userID, id int64) error {
	ok, err := s.store.SoftDeleteAlert(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("delete alert: %w", err)
	}
	if !ok {
		return ErrNotFound
	}
	return nil
}

// Evaluate walks active alerts in batches, records unique events, then
// attempts delivery. A send failure keeps the alert and retries later.
func (s *Service) Evaluate(ctx context.Context) error {
	alerts, err := s.store.ListActiveAlertsForEval(ctx)
	if err != nil {
		return fmt.Errorf("list alerts: %w", err)
	}
	ids := make([]int64, 0, len(alerts))
	seen := map[int64]struct{}{}
	for _, a := range alerts {
		if _, ok := seen[a.ProductID]; ok {
			continue
		}
		seen[a.ProductID] = struct{}{}
		ids = append(ids, a.ProductID)
	}
	quotes, err := s.store.ListProductQuotes(ctx, ids, s.freshSince())
	if err != nil {
		return fmt.Errorf("list quotes: %w", err)
	}
	byID := make(map[int64]Quote, len(quotes))
	for _, q := range quotes {
		byID[q.ProductID] = q
	}
	for _, a := range alerts {
		q := byID[a.ProductID]
		if err := s.apply(ctx, a, q); err != nil {
			return err
		}
	}
	return s.deliver(ctx)
}

func (s *Service) apply(ctx context.Context, a evalAlert, q Quote) error {
	switch a.Kind {
	case KindPriceDrop:
		if ShouldFirePriceDrop(a.BaselinePriceRial, a.TargetPriceRial, q.LowestInStockRial) {
			key := DropEventKey(q.LowestInStockRial)
			inserted, err := s.store.InsertAlertDelivery(ctx, a.ID, key)
			if err != nil {
				return fmt.Errorf("insert delivery: %w", err)
			}
			if inserted {
				s.log.InfoContext(ctx, "alert.triggered",
					"alert_id", a.ID, "product_id", a.ProductID, "event", key,
					"phone", logger.MaskPhone(a.Phone))
			}
			if err := s.store.UpdateAlertBaseline(ctx, a.ID, q.LowestInStockRial); err != nil {
				return fmt.Errorf("update baseline: %w", err)
			}
		} else if q.LowestInStockRial > a.BaselinePriceRial {
			if err := s.store.UpdateAlertBaseline(ctx, a.ID, q.LowestInStockRial); err != nil {
				return fmt.Errorf("update baseline: %w", err)
			}
		}
	case KindBackInStock:
		if ShouldFireBackInStock(a.StockWasAvailable, q.InStock) {
			inserted, err := s.store.InsertAlertDelivery(ctx, a.ID, stockEventKey)
			if err != nil {
				return fmt.Errorf("insert delivery: %w", err)
			}
			if inserted {
				s.log.InfoContext(ctx, "alert.triggered",
					"alert_id", a.ID, "product_id", a.ProductID, "event", stockEventKey,
					"phone", logger.MaskPhone(a.Phone))
			}
		}
		if a.StockWasAvailable != q.InStock {
			if err := s.store.UpdateAlertStockFlag(ctx, a.ID, q.InStock); err != nil {
				return fmt.Errorf("update stock flag: %w", err)
			}
		}
	}
	return nil
}

func (s *Service) deliver(ctx context.Context) error {
	items, err := s.store.ListRetryableDeliveries(ctx)
	if err != nil {
		return fmt.Errorf("list deliveries: %w", err)
	}
	for _, d := range items {
		msg := deliveryMessage(d.Kind, d.ProductName)
		if err := s.sms.Notify(d.Phone, msg); err != nil {
			s.log.WarnContext(ctx, "alert.send_failed",
				"alert_id", d.AlertID, "product_id", d.ProductID,
				"phone", logger.MaskPhone(d.Phone), "error", err.Error())
			if markErr := s.store.MarkDeliveryFailed(ctx, d.ID, err.Error()); markErr != nil {
				return fmt.Errorf("mark failed: %w", markErr)
			}
			continue
		}
		if err := s.store.MarkDeliverySent(ctx, d.ID); err != nil {
			return fmt.Errorf("mark sent: %w", err)
		}
		s.log.InfoContext(ctx, "alert.sent",
			"alert_id", d.AlertID, "product_id", d.ProductID,
			"phone", logger.MaskPhone(d.Phone))
	}
	return nil
}

func (s *Service) freshSince() time.Time {
	return s.now().UTC().Add(-s.freshAfter)
}

func deliveryMessage(kind, name string) string {
	if kind == KindBackInStock {
		return "«" + name + "» دوباره موجود شد."
	}
	return "قیمت «" + name + "» کاهش یافت."
}
