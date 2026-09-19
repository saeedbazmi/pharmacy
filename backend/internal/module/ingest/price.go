package ingest

// IsSuspiciousJump reports whether a new price should be withheld from the
// public site. The function is pure: no I/O, no time, no logging.
//
// Rules:
//   - first observation (old <= 0) is never a jump
//   - a drop to zero from a real price is always suspicious
//   - otherwise the relative change must exceed threshold (e.g. 0.70 = ±70%)
func IsSuspiciousJump(oldPrice, newPrice int64, threshold float64) bool {
	if threshold <= 0 {
		return false
	}
	if oldPrice <= 0 {
		return false
	}
	if newPrice <= 0 {
		return true
	}
	if oldPrice == newPrice {
		return false
	}
	delta := float64(newPrice-oldPrice) / float64(oldPrice)
	if delta < 0 {
		delta = -delta
	}
	return delta > threshold
}
