package search

import (
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

func parseQuery(values url.Values) (Query, error) {
	raw := values.Get("q")
	if utf8.RuneCountInString(raw) > textfa.MaxQueryRunes*2 {
		return Query{}, ErrQueryTooLong
	}
	q := textfa.Query(raw)
	if q == "" {
		return Query{}, ErrQueryEmpty
	}
	sort := SortRelevance
	switch values.Get("sort") {
	case SortPriceAsc:
		sort = SortPriceAsc
	case SortPriceDesc:
		sort = SortPriceDesc
	}
	inStock := false
	switch values.Get("in_stock") {
	case "1", "true", "yes":
		inStock = true
	}
	page := DefaultPage
	if n, err := strconv.Atoi(values.Get("page")); err == nil && n > 0 {
		page = n
	}
	size := DefaultSize
	if n, err := strconv.Atoi(values.Get("page_size")); err == nil && n > 0 {
		size = n
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	return Query{
		Q:        q,
		Brand:    safeSlug(values.Get("brand")),
		Category: safeSlug(values.Get("category")),
		InStock:  inStock,
		Sort:     sort,
		Page:     page,
		PageSize: size,
	}, nil
}

func safeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" || len(s) > 80 {
		return ""
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
		default:
			return ""
		}
	}
	return s
}

func cacheKey(q Query) string {
	stock := "0"
	if q.InStock {
		stock = "1"
	}
	return strings.Join([]string{q.Q, q.Brand, q.Category, stock, q.Sort, strconv.Itoa(q.Page), strconv.Itoa(q.PageSize)}, "|")
}

// rankBucket mirrors the SQL ranking: exact, prefix, document-prefix, trigram.
func rankBucket(q, nameNormalized, document string) int {
	if nameNormalized == q {
		return 0
	}
	if strings.HasPrefix(nameNormalized, q) {
		return 1
	}
	if strings.HasPrefix(document, q) {
		return 2
	}
	return 3
}
