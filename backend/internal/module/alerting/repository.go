package alerting

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

type Store interface {
	UpsertFavorite(ctx context.Context, userID, productID int64) error
	DeleteFavorite(ctx context.Context, userID, productID int64) error
	FavoriteExists(ctx context.Context, userID, productID int64) (bool, error)
	ListFavorites(ctx context.Context, userID int64, freshSince time.Time) ([]Favorite, error)
	ProductPublished(ctx context.Context, productID int64) (bool, error)
	InsertAlert(ctx context.Context, userID int64, in AlertInput, baseline int64, stock bool) (Alert, error)
	CountActiveAlerts(ctx context.Context, userID int64) (int64, error)
	ListUserAlerts(ctx context.Context, userID int64) ([]Alert, error)
	SoftDeleteAlert(ctx context.Context, id, userID int64) (bool, error)
	ListActiveAlertsForEval(ctx context.Context) ([]evalAlert, error)
	ListProductQuotes(ctx context.Context, ids []int64, freshSince time.Time) ([]Quote, error)
	UpdateAlertBaseline(ctx context.Context, id, baseline int64) error
	UpdateAlertStockFlag(ctx context.Context, id int64, available bool) error
	InsertAlertDelivery(ctx context.Context, alertID int64, eventKey string) (bool, error)
	ListRetryableDeliveries(ctx context.Context) ([]delivery, error)
	MarkDeliverySent(ctx context.Context, id int64) error
	MarkDeliveryFailed(ctx context.Context, id int64, lastError string) error
}

type repository struct {
	q *store.Queries
}

func NewStore(pool *pgxpool.Pool) Store {
	return &repository{q: store.New(pool)}
}

func (r *repository) UpsertFavorite(ctx context.Context, userID, productID int64) error {
	return r.q.UpsertFavorite(ctx, store.UpsertFavoriteParams{UserID: userID, ProductID: productID})
}

func (r *repository) DeleteFavorite(ctx context.Context, userID, productID int64) error {
	return r.q.DeleteFavorite(ctx, store.DeleteFavoriteParams{UserID: userID, ProductID: productID})
}

func (r *repository) FavoriteExists(ctx context.Context, userID, productID int64) (bool, error) {
	return r.q.FavoriteExists(ctx, store.FavoriteExistsParams{UserID: userID, ProductID: productID})
}

func (r *repository) ListFavorites(ctx context.Context, userID int64, freshSince time.Time) ([]Favorite, error) {
	rows, err := r.q.ListFavorites(ctx, store.ListFavoritesParams{UserID: userID, FreshSince: freshSince})
	if err != nil {
		return nil, err
	}
	out := make([]Favorite, 0, len(rows))
	for _, row := range rows {
		item := Favorite{
			ProductID: row.ID, Slug: row.Slug, NameFa: row.NameFa,
			LowestPriceRial: row.LowestPriceRial, OfferCount: int(row.OfferCount), CreatedAt: row.CreatedAt,
		}
		if row.ImageUrl != nil {
			item.ImageURL = *row.ImageUrl
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *repository) ProductPublished(ctx context.Context, productID int64) (bool, error) {
	return r.q.PublishedProductExists(ctx, productID)
}

func (r *repository) InsertAlert(ctx context.Context, userID int64, in AlertInput, baseline int64, stock bool) (Alert, error) {
	row, err := r.q.InsertPriceAlert(ctx, store.InsertPriceAlertParams{
		UserID:            userID,
		ProductID:         in.ProductID,
		Kind:              in.Kind,
		TargetPriceRial:   in.TargetPriceRial,
		BaselinePriceRial: baseline,
		StockWasAvailable: stock,
	})
	if err != nil {
		return Alert{}, err
	}
	return Alert{
		ID: row.ID, UserID: row.UserID, ProductID: row.ProductID, Kind: row.Kind,
		TargetPriceRial: row.TargetPriceRial, BaselinePriceRial: row.BaselinePriceRial,
		StockWasAvailable: row.StockWasAvailable, Status: row.Status, CreatedAt: row.CreatedAt,
	}, nil
}

func (r *repository) CountActiveAlerts(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountActiveAlerts(ctx, userID)
}

func (r *repository) ListUserAlerts(ctx context.Context, userID int64) ([]Alert, error) {
	rows, err := r.q.ListUserAlerts(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Alert, 0, len(rows))
	for _, row := range rows {
		out = append(out, Alert{
			ID: row.ID, UserID: row.UserID, ProductID: row.ProductID, Kind: row.Kind,
			TargetPriceRial: row.TargetPriceRial, BaselinePriceRial: row.BaselinePriceRial,
			StockWasAvailable: row.StockWasAvailable, Status: row.Status, CreatedAt: row.CreatedAt,
			ProductSlug: row.ProductSlug, ProductName: row.ProductName,
		})
	}
	return out, nil
}

func (r *repository) SoftDeleteAlert(ctx context.Context, id, userID int64) (bool, error) {
	n, err := r.q.SoftDeleteAlert(ctx, store.SoftDeleteAlertParams{ID: id, UserID: userID})
	return n > 0, err
}

func (r *repository) ListActiveAlertsForEval(ctx context.Context) ([]evalAlert, error) {
	rows, err := r.q.ListActiveAlertsForEval(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]evalAlert, 0, len(rows))
	for _, row := range rows {
		out = append(out, evalAlert{
			Alert: Alert{
				ID: row.ID, UserID: row.UserID, ProductID: row.ProductID, Kind: row.Kind,
				TargetPriceRial: row.TargetPriceRial, BaselinePriceRial: row.BaselinePriceRial,
				StockWasAvailable: row.StockWasAvailable,
			},
			Phone: row.Phone, ProductName: row.ProductName,
		})
	}
	return out, nil
}

func (r *repository) ListProductQuotes(ctx context.Context, ids []int64, freshSince time.Time) ([]Quote, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.q.ListProductQuotes(ctx, store.ListProductQuotesParams{FreshSince: freshSince, ProductIds: ids})
	if err != nil {
		return nil, err
	}
	out := make([]Quote, 0, len(rows))
	for _, row := range rows {
		out = append(out, Quote{ProductID: row.ProductID, LowestInStockRial: row.LowestInStockRial, InStock: row.InStock})
	}
	return out, nil
}

func (r *repository) UpdateAlertBaseline(ctx context.Context, id, baseline int64) error {
	return r.q.UpdateAlertBaseline(ctx, store.UpdateAlertBaselineParams{ID: id, BaselinePriceRial: baseline})
}

func (r *repository) UpdateAlertStockFlag(ctx context.Context, id int64, available bool) error {
	return r.q.UpdateAlertStockFlag(ctx, store.UpdateAlertStockFlagParams{ID: id, StockWasAvailable: available})
}

func (r *repository) InsertAlertDelivery(ctx context.Context, alertID int64, eventKey string) (bool, error) {
	n, err := r.q.InsertAlertDelivery(ctx, store.InsertAlertDeliveryParams{AlertID: alertID, EventKey: eventKey})
	return n > 0, err
}

func (r *repository) ListRetryableDeliveries(ctx context.Context) ([]delivery, error) {
	rows, err := r.q.ListRetryableDeliveries(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]delivery, 0, len(rows))
	for _, row := range rows {
		out = append(out, delivery{
			ID: row.ID, AlertID: row.AlertID, EventKey: row.EventKey, Attempts: row.Attempts,
			UserID: row.UserID, ProductID: row.ProductID, Kind: row.Kind, Phone: row.Phone,
			ProductName: row.ProductName,
		})
	}
	return out, nil
}

func (r *repository) MarkDeliverySent(ctx context.Context, id int64) error {
	return r.q.MarkDeliverySent(ctx, id)
}

func (r *repository) MarkDeliveryFailed(ctx context.Context, id int64, lastError string) error {
	return r.q.MarkDeliveryFailed(ctx, store.MarkDeliveryFailedParams{ID: id, LastError: &lastError})
}

var _ Store = (*repository)(nil)
