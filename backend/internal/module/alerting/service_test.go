package alerting

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

type memStore struct {
	published  map[int64]bool
	favorites  map[[2]int64]Favorite
	alerts     map[int64]evalAlert
	quotes     map[int64]Quote
	deliveries map[string]delivery
	nextID     int64
	sent       []int64
	failed     []int64
}

func newMem() *memStore {
	return &memStore{
		published:  map[int64]bool{1: true},
		favorites:  map[[2]int64]Favorite{},
		alerts:     map[int64]evalAlert{},
		quotes:     map[int64]Quote{},
		deliveries: map[string]delivery{},
		nextID:     1,
	}
}

func (m *memStore) UpsertFavorite(_ context.Context, userID, productID int64) error {
	m.favorites[[2]int64{userID, productID}] = Favorite{ProductID: productID, Slug: "p", NameFa: "کالا"}
	return nil
}
func (m *memStore) DeleteFavorite(_ context.Context, userID, productID int64) error {
	delete(m.favorites, [2]int64{userID, productID})
	return nil
}
func (m *memStore) FavoriteExists(_ context.Context, userID, productID int64) (bool, error) {
	_, ok := m.favorites[[2]int64{userID, productID}]
	return ok, nil
}
func (m *memStore) ListFavorites(_ context.Context, userID int64, _ time.Time) ([]Favorite, error) {
	var out []Favorite
	for key, fav := range m.favorites {
		if key[0] == userID {
			out = append(out, fav)
		}
	}
	return out, nil
}
func (m *memStore) ProductPublished(_ context.Context, productID int64) (bool, error) {
	return m.published[productID], nil
}
func (m *memStore) InsertAlert(_ context.Context, userID int64, in AlertInput, baseline int64, stock bool) (Alert, error) {
	id := m.nextID
	m.nextID++
	a := evalAlert{
		Alert: Alert{
			ID: id, UserID: userID, ProductID: in.ProductID, Kind: in.Kind,
			TargetPriceRial: in.TargetPriceRial, BaselinePriceRial: baseline,
			StockWasAvailable: stock, Status: "active", ProductName: "کالا",
		},
		Phone: "09121234567", ProductName: "کالا",
	}
	m.alerts[id] = a
	return a.Alert, nil
}
func (m *memStore) CountActiveAlerts(_ context.Context, userID int64) (int64, error) {
	var n int64
	for _, a := range m.alerts {
		if a.UserID == userID && a.Status == "active" {
			n++
		}
	}
	return n, nil
}
func (m *memStore) ListUserAlerts(context.Context, int64) ([]Alert, error) { return nil, nil }
func (m *memStore) SoftDeleteAlert(_ context.Context, id, userID int64) (bool, error) {
	a, ok := m.alerts[id]
	if !ok || a.UserID != userID {
		return false, nil
	}
	a.Status = "paused"
	m.alerts[id] = a
	return true, nil
}
func (m *memStore) ListActiveAlertsForEval(context.Context) ([]evalAlert, error) {
	var out []evalAlert
	for _, a := range m.alerts {
		if a.Status == "active" {
			out = append(out, a)
		}
	}
	return out, nil
}
func (m *memStore) ListProductQuotes(_ context.Context, ids []int64, _ time.Time) ([]Quote, error) {
	var out []Quote
	for _, id := range ids {
		if q, ok := m.quotes[id]; ok {
			out = append(out, q)
		}
	}
	return out, nil
}
func (m *memStore) UpdateAlertBaseline(_ context.Context, id, baseline int64) error {
	a := m.alerts[id]
	a.BaselinePriceRial = baseline
	m.alerts[id] = a
	return nil
}
func (m *memStore) UpdateAlertStockFlag(_ context.Context, id int64, available bool) error {
	a := m.alerts[id]
	a.StockWasAvailable = available
	m.alerts[id] = a
	return nil
}
func (m *memStore) InsertAlertDelivery(_ context.Context, alertID int64, eventKey string) (bool, error) {
	key := strconvKey(alertID, eventKey)
	if _, ok := m.deliveries[key]; ok {
		return false, nil
	}
	id := m.nextID
	m.nextID++
	a := m.alerts[alertID]
	m.deliveries[key] = delivery{
		ID: id, AlertID: alertID, EventKey: eventKey, Phone: a.Phone,
		ProductID: a.ProductID, Kind: a.Kind, ProductName: a.ProductName, UserID: a.UserID,
	}
	return true, nil
}
func (m *memStore) ListRetryableDeliveries(context.Context) ([]delivery, error) {
	var out []delivery
	for _, d := range m.deliveries {
		if d.Attempts < 5 {
			out = append(out, d)
		}
	}
	return out, nil
}
func (m *memStore) MarkDeliverySent(_ context.Context, id int64) error {
	m.sent = append(m.sent, id)
	for key, d := range m.deliveries {
		if d.ID == id {
			d.Attempts = 99
			m.deliveries[key] = d
		}
	}
	return nil
}
func (m *memStore) MarkDeliveryFailed(_ context.Context, id int64, _ string) error {
	m.failed = append(m.failed, id)
	for key, d := range m.deliveries {
		if d.ID == id {
			d.Attempts++
			m.deliveries[key] = d
		}
	}
	return nil
}

func strconvKey(id int64, key string) string {
	return strings.Join([]string{itoa(id), key}, "|")
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

type captureNotify struct {
	n     int
	phone string
	msg   string
	err   error
}

func (c *captureNotify) Notify(phone, message string) error {
	c.n++
	c.phone, c.msg = phone, message
	return c.err
}

func TestAddFavoriteIsIdempotent(t *testing.T) {
	store := newMem()
	svc := NewService(store, &captureNotify{}, time.Hour, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	if err := svc.AddFavorite(context.Background(), 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddFavorite(context.Background(), 1, 1); err != nil {
		t.Fatal(err)
	}
}

func TestCreateAlertEnforcesCap(t *testing.T) {
	store := newMem()
	svc := NewService(store, &captureNotify{}, time.Hour, nil)
	for i := 0; i < MaxActiveAlerts; i++ {
		if _, err := svc.CreateAlert(context.Background(), 9, AlertInput{ProductID: 1, Kind: KindPriceDrop}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CreateAlert(context.Background(), 9, AlertInput{ProductID: 1, Kind: KindPriceDrop}); err != ErrAlertLimit {
		t.Fatalf("err = %v", err)
	}
}

func TestEvaluatePriceDropNotifiesOnce(t *testing.T) {
	store := newMem()
	notify := &captureNotify{}
	var logs bytes.Buffer
	svc := NewService(store, notify, time.Hour, slog.New(slog.NewJSONHandler(&logs, nil)))
	alert, err := svc.CreateAlert(context.Background(), 1, AlertInput{ProductID: 1, Kind: KindPriceDrop})
	if err != nil {
		t.Fatal(err)
	}
	a := store.alerts[alert.ID]
	a.BaselinePriceRial = 100000
	a.Phone = "09121234567"
	store.alerts[alert.ID] = a
	store.quotes[1] = Quote{ProductID: 1, LowestInStockRial: 80000, InStock: true}

	if err := svc.Evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if notify.n != 1 {
		t.Fatalf("sends = %d, want 1", notify.n)
	}
	if err := svc.Evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if notify.n != 1 {
		t.Fatalf("second eval sends = %d, want still 1", notify.n)
	}
	body := logs.String()
	if !strings.Contains(body, "alert.triggered") || !strings.Contains(body, "alert.sent") {
		t.Fatalf("missing events: %s", body)
	}
	if strings.Contains(body, "09121234567") {
		t.Fatal("full phone leaked into logs")
	}
}

func TestEvaluateSkipsMissingQuote(t *testing.T) {
	store := newMem()
	notify := &captureNotify{}
	svc := NewService(store, notify, time.Hour, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	alert, err := svc.CreateAlert(context.Background(), 1, AlertInput{ProductID: 1, Kind: KindPriceDrop})
	if err != nil {
		t.Fatal(err)
	}
	a := store.alerts[alert.ID]
	a.BaselinePriceRial = 100000
	store.alerts[alert.ID] = a
	store.quotes[1] = Quote{ProductID: 1, LowestInStockRial: 0, InStock: false}
	if err := svc.Evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if notify.n != 0 {
		t.Fatalf("sends = %d, stale/empty quote must not notify", notify.n)
	}
}

func TestSendFailureKeepsAlert(t *testing.T) {
	store := newMem()
	notify := &captureNotify{err: ErrUnauthorized}
	svc := NewService(store, notify, time.Hour, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	alert, err := svc.CreateAlert(context.Background(), 1, AlertInput{ProductID: 1, Kind: KindPriceDrop})
	if err != nil {
		t.Fatal(err)
	}
	a := store.alerts[alert.ID]
	a.BaselinePriceRial = 100000
	store.alerts[alert.ID] = a
	store.quotes[1] = Quote{ProductID: 1, LowestInStockRial: 50000, InStock: true}
	if err := svc.Evaluate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.alerts[alert.ID].Status != "active" {
		t.Fatal("failed send must not delete the alert")
	}
	if len(store.failed) != 1 {
		t.Fatalf("failed marks = %d", len(store.failed))
	}
}
