package identity

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/store"
)

type Store interface {
	GetUserByPhone(ctx context.Context, phone string) (User, error)
	GetUserByID(ctx context.Context, id int64) (User, error)
	InsertUser(ctx context.Context, phone string) (User, error)
	InsertOTP(ctx context.Context, phone, hash, ip string, expires time.Time) (int64, error)
	LatestOTP(ctx context.Context, phone string) (otpRecord, error)
	ConsumeOTP(ctx context.Context, id int64) (bool, error)
	BumpOTPAttempts(ctx context.Context, id int64) error
	InsertOTPSend(ctx context.Context, phone, ip string) error
	CountOTPSendsPhone(ctx context.Context, phone string, since time.Time) (int64, error)
	CountOTPSendsIP(ctx context.Context, ip string, since time.Time) (int64, error)
	InsertSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error
	UserBySession(ctx context.Context, tokenHash string) (User, time.Time, error)
	DeleteSession(ctx context.Context, tokenHash string) error
	ListSearches(ctx context.Context, userID int64) ([]SearchEntry, error)
	DeleteSearches(ctx context.Context, userID int64) error
}

type SearchEntry struct {
	ID      int64
	At      time.Time
	Query   string
	Results int32
}

type repository struct {
	q *store.Queries
}

func NewStore(pool *pgxpool.Pool) Store {
	return &repository{q: store.New(pool)}
}

func (r *repository) GetUserByPhone(ctx context.Context, phone string) (User, error) {
	row, err := r.q.GetUserByPhone(ctx, phone)
	if err != nil {
		return User{}, mapMissing(err)
	}
	return User{ID: row.ID, Phone: row.Phone, Role: row.Role}, nil
}

func (r *repository) GetUserByID(ctx context.Context, id int64) (User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return User{}, mapMissing(err)
	}
	return User{ID: row.ID, Phone: row.Phone, Role: row.Role}, nil
}

func (r *repository) InsertUser(ctx context.Context, phone string) (User, error) {
	row, err := r.q.InsertUser(ctx, phone)
	if err != nil {
		return User{}, err
	}
	return User{ID: row.ID, Phone: row.Phone, Role: row.Role}, nil
}

func (r *repository) InsertOTP(ctx context.Context, phone, hash, ip string, expires time.Time) (int64, error) {
	return r.q.InsertOTPCode(ctx, store.InsertOTPCodeParams{
		Phone: phone, CodeHash: hash, ExpiresAt: expires, Ip: ip,
	})
}

func (r *repository) LatestOTP(ctx context.Context, phone string) (otpRecord, error) {
	row, err := r.q.GetLatestOTP(ctx, phone)
	if err != nil {
		return otpRecord{}, mapMissing(err)
	}
	return otpRecord{
		ID: row.ID, Phone: row.Phone, Hash: row.CodeHash, ExpiresAt: row.ExpiresAt,
		Consumed: row.ConsumedAt != nil, Attempts: row.AttemptCount, CreatedAt: row.CreatedAt,
	}, nil
}

func (r *repository) ConsumeOTP(ctx context.Context, id int64) (bool, error) {
	n, err := r.q.ConsumeOTP(ctx, id)
	return n > 0, err
}

func (r *repository) BumpOTPAttempts(ctx context.Context, id int64) error {
	return r.q.IncrementOTPAttempts(ctx, id)
}

func (r *repository) InsertOTPSend(ctx context.Context, phone, ip string) error {
	return r.q.InsertOTPSend(ctx, store.InsertOTPSendParams{Phone: phone, Ip: ip})
}

func (r *repository) CountOTPSendsPhone(ctx context.Context, phone string, since time.Time) (int64, error) {
	return r.q.CountOTPSendsByPhone(ctx, store.CountOTPSendsByPhoneParams{Phone: phone, CreatedAt: since})
}

func (r *repository) CountOTPSendsIP(ctx context.Context, ip string, since time.Time) (int64, error) {
	return r.q.CountOTPSendsByIP(ctx, store.CountOTPSendsByIPParams{Ip: ip, CreatedAt: since})
}

func (r *repository) InsertSession(ctx context.Context, userID int64, tokenHash string, expires time.Time) error {
	return r.q.InsertUserSession(ctx, store.InsertUserSessionParams{
		UserID: userID, TokenHash: tokenHash, ExpiresAt: expires,
	})
}

func (r *repository) UserBySession(ctx context.Context, tokenHash string) (User, time.Time, error) {
	row, err := r.q.GetUserBySessionHash(ctx, tokenHash)
	if err != nil {
		return User{}, time.Time{}, ErrUnauthorized
	}
	return User{ID: row.ID, Phone: row.Phone, Role: row.Role}, row.ExpiresAt, nil
}

func (r *repository) DeleteSession(ctx context.Context, tokenHash string) error {
	return r.q.DeleteUserSession(ctx, tokenHash)
}

func (r *repository) ListSearches(ctx context.Context, userID int64) ([]SearchEntry, error) {
	rows, err := r.q.ListUserSearches(ctx, &userID)
	if err != nil {
		return nil, err
	}
	out := make([]SearchEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, SearchEntry{ID: row.ID, At: row.QueriedAt, Query: row.QueryNormalized, Results: row.ResultCount})
	}
	return out, nil
}

func (r *repository) DeleteSearches(ctx context.Context, userID int64) error {
	return r.q.DeleteUserSearches(ctx, &userID)
}

func mapMissing(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

var _ Store = (*repository)(nil)
