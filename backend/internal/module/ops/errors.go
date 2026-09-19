package ops

import "errors"

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidLogin    = errors.New("invalid login")
	ErrRateLimited     = errors.New("login rate limited")
	ErrInactiveUser    = errors.New("inactive user")
	ErrInvalidInput    = errors.New("invalid input")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrHasChildren     = errors.New("has children")
	ErrHasProducts     = errors.New("has products")
	ErrUnknownFetcher  = errors.New("unknown fetcher")
	ErrUnsafeURL       = errors.New("unsafe url")
	ErrInvalidSchedule = errors.New("invalid schedule")
	ErrInvalidConfig   = errors.New("invalid source config")
	ErrMatchNotPending = errors.New("match is not pending")
	ErrMatchNotDecided = errors.New("match is not decided")
	ErrSameBrand       = errors.New("cannot merge a brand into itself")
	ErrNotSuspicious   = errors.New("offer is not awaiting price review")
)
