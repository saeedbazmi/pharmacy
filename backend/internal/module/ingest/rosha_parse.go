package ingest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	roshaTitleRe = regexp.MustCompile(`<a class="title overflow-hidden d-block" title="([^"]*)" href="(https://roshapharmacy\.com/product/(\d+))"\s*>([^<]*)</a>`)
	roshaPriceRe = regexp.MustCompile(`<span class="price">([0-9۰-۹٠-٩,٬]+)</span>`)
	roshaImgRe   = regexp.MustCompile(`data-src="(https://[^"]+)"`)
	roshaImgSrc  = regexp.MustCompile(`<img[^>]+src="(https://[^"]+)"`)
)

type roshaRaw struct {
	ExternalID string `json:"external_id"`
	NameFa     string `json:"name_fa"`
	ProductURL string `json:"product_url"`
	ImageURL   string `json:"image_url,omitempty"`
	PriceToman int64  `json:"price_toman"`
	InStock    bool   `json:"in_stock"`
}

// ParseRoshaListing extracts product cards from a Rosha category or home page.
func ParseRoshaListing(body []byte) ([]RawItem, error) {
	s := string(body)
	matches := roshaTitleRe.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("rosha listing contained no products")
	}
	items := make([]RawItem, 0, len(matches))
	seen := make(map[string]bool, len(matches))
	for _, loc := range matches {
		title := htmlUnescape(s[loc[8]:loc[9]])
		url := s[loc[4]:loc[5]]
		id := s[loc[6]:loc[7]]
		if title == "" || seen[id] {
			continue
		}
		seen[id] = true
		end := loc[1] + 1800
		if end > len(s) {
			end = len(s)
		}
		window := s[loc[0]:end]
		price := parseToman(firstSub(roshaPriceRe, window))
		img := firstSub(roshaImgRe, window)
		if img == "" || strings.Contains(img, "rosha-null") {
			img = firstSub(roshaImgSrc, window)
		}
		if strings.Contains(img, "rosha-null") {
			img = ""
		}
		inStock := strings.Contains(window, "btn-basket")
		raw, err := json.Marshal(roshaRaw{
			ExternalID: id,
			NameFa:     strings.TrimSpace(title),
			ProductURL: url,
			ImageURL:   img,
			PriceToman: price,
			InStock:    inStock,
		})
		if err != nil {
			continue
		}
		items = append(items, RawItem{
			ExternalID: id,
			NameFa:     strings.TrimSpace(title),
			ImageURL:   img,
			ProductURL: url,
			PriceToman: price,
			InStock:    inStock,
			SlugHint:   "rosha-" + id,
			Raw:        raw,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("rosha listing produced no usable items")
	}
	return items, nil
}

func firstSub(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func parseToman(raw string) int64 {
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return 0
	}
	n, err := strconv.ParseInt(b.String(), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func htmlUnescape(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&#39;", "'")
	s = strings.ReplaceAll(s, "&quot;", `"`)
	return s
}
