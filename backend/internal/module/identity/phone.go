package identity

import (
	"regexp"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/logger"
	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

var irMobile = regexp.MustCompile(`^09[0-9]{9}$`)

// NormalizePhone accepts common Iranian writings and returns 09xxxxxxxxx.
func NormalizePhone(raw string) (string, error) {
	s := textfa.Normalize(raw)
	s = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, s)
	switch {
	case strings.HasPrefix(s, "0098") && len(s) == 14:
		s = "0" + s[4:]
	case strings.HasPrefix(s, "98") && len(s) == 12:
		s = "0" + s[2:]
	case strings.HasPrefix(s, "9") && len(s) == 10:
		s = "0" + s
	}
	if !irMobile.MatchString(s) {
		return "", ErrInvalidPhone
	}
	return s, nil
}

// MaskPhone hides the subscriber for logs and API responses.
func MaskPhone(phone string) string {
	return logger.MaskPhone(phone)
}
