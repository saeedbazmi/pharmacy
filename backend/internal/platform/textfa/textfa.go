package textfa

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

var digitReplacer = strings.NewReplacer(
	"۰", "0", "۱", "1", "۲", "2", "۳", "3", "۴", "4",
	"۵", "5", "۶", "6", "۷", "7", "۸", "8", "۹", "9",
	"٠", "0", "١", "1", "٢", "2", "٣", "3", "٤", "4",
	"٥", "5", "٦", "6", "٧", "7", "٨", "8", "٩", "9",
)

// Normalize applies the shared Persian rules used by both ingest matching and
// search: Arabic ye/kaf become Persian, digits become Latin, zero-width marks
// are dropped, and ZWNJ becomes a space.
func Normalize(input string) string {
	if input == "" {
		return ""
	}
	s := digitReplacer.Replace(input)

	var b strings.Builder
	b.Grow(len(s))
	prevSpace := true
	for _, r := range s {
		switch r {
		case 'ي', 'ى':
			r = 'ی'
		case 'ك':
			r = 'ک'
		case 'أ', 'إ', 'آ':
			r = 'ا'
		case 'ة':
			r = 'ه'
		case '\u200c': // ZWNJ
			r = ' '
		case '\u200b', '\u200d', '\u200e', '\u200f', '\ufeff':
			continue
		}
		if unicode.Is(unicode.Mn, r) { // harakat
			continue
		}
		if unicode.IsSpace(r) {
			if prevSpace {
				continue
			}
			b.WriteByte(' ')
			prevSpace = true
			continue
		}
		if unicode.IsControl(r) {
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// Slugify turns a latin-ish identifier into a URL slug: lowercase, digits and
// hyphens only. Persian letters are dropped so the slug stays ASCII and stable.
func Slugify(input string) string {
	s := strings.ToLower(Normalize(input))
	var b strings.Builder
	b.Grow(len(s))
	lastHyphen := true
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

const MaxQueryRunes = 120

// Query prepares a user search string: same rules as Normalize, then lower-cased
// and capped. Empty after normalisation means the caller should reject it.
func Query(input string) string {
	s := strings.ToLower(Normalize(input))
	if s == "" {
		return ""
	}
	if utf8.RuneCountInString(s) <= MaxQueryRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:MaxQueryRunes])
}

// Document concatenates catalogue fields into the stored search blob.
func Document(parts ...string) string {
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, part := range parts {
		n := strings.ToLower(Normalize(part))
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	return strings.Join(out, " ")
}

// IsBlank reports whether s is empty after normalisation.
func IsBlank(s string) bool {
	return Normalize(s) == ""
}
