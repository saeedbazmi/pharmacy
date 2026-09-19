package alerting

import "strconv"

// DropEventKey is unique per alert and observed price so a later drop to a
// new amount can notify again, while the same amount cannot.
func DropEventKey(price int64) string {
	return "drop:" + strconv.FormatInt(price, 10)
}

// ShouldFirePriceDrop reports whether the current valid in-stock price is a
// drop the user asked to hear about.
func ShouldFirePriceDrop(baseline int64, target *int64, current int64) bool {
	if current <= 0 {
		return false
	}
	if current >= baseline {
		return false
	}
	if target != nil && current > *target {
		return false
	}
	return true
}

// ShouldFireBackInStock reports a false → true availability transition.
func ShouldFireBackInStock(wasAvailable, nowInStock bool) bool {
	return !wasAvailable && nowInStock
}
