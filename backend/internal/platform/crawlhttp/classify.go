package crawlhttp

import (
	"context"
	"errors"
	"net"
)

type classifiedError struct {
	err       error
	transient bool
}

func (e classifiedError) Error() string { return e.err.Error() }
func (e classifiedError) Unwrap() error { return e.err }

// Transient wraps a network or 5xx failure that is worth retrying.
func Transient(err error) error {
	if err == nil {
		return nil
	}
	return classifiedError{err: err, transient: true}
}

// Structural wraps a parse, 4xx, or robots failure that will not heal on retry.
func Structural(err error) error {
	if err == nil {
		return nil
	}
	return classifiedError{err: err, transient: false}
}

// IsTransient reports a retryable network/status failure.
func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) {
		return false
	}
	var c classifiedError
	if errors.As(err, &c) {
		return c.transient
	}
	var netErr net.Error
	return errors.As(err, &netErr)
}

// IsStructural reports a failure that should not burn through retries.
func IsStructural(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	var c classifiedError
	if errors.As(err, &c) {
		return !c.transient
	}
	return !IsTransient(err)
}
