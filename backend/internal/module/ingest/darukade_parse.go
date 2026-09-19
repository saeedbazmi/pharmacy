package ingest

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

type listingPage struct {
	Props struct {
		PageProps struct {
			CategoryMataData struct {
				Products []listingProduct `json:"products"`
				Filters  struct {
					Pages struct {
						TotalPage   int `json:"totalPage"`
						CurrentPage int `json:"currentPage"`
					} `json:"pages"`
				} `json:"filters"`
			} `json:"categoryMataData"`
		} `json:"pageProps"`
	} `json:"props"`
}

type listingProduct struct {
	ID           int64  `json:"id"`
	FaName       string `json:"faName"`
	ProductURL   string `json:"productURL"`
	Image        string `json:"image"`
	BrandEnName  string `json:"brandEnName"`
	OrginalCost  int64  `json:"orginalCost"`
	DiscountCost int64  `json:"discountCost"`
	StoreState   bool   `json:"storeState"`
	StoreCount   int    `json:"storeCount"`
}

// ParseListing extracts products from a Darukade category page. The body may be
// raw HTML containing __NEXT_DATA__ or the JSON object itself.
func ParseListing(body []byte, origin string) ([]RawItem, error) {
	payload, err := nextData(body)
	if err != nil {
		return nil, err
	}
	var page listingPage
	if err := json.Unmarshal(payload, &page); err != nil {
		return nil, fmt.Errorf("decode darukade listing: %w", err)
	}
	products := page.Props.PageProps.CategoryMataData.Products
	if len(products) == 0 {
		return nil, fmt.Errorf("darukade listing contained no products")
	}
	items := make([]RawItem, 0, len(products))
	for _, p := range products {
		if p.ID == 0 || p.FaName == "" {
			continue
		}
		raw, _ := json.Marshal(p)
		price := p.DiscountCost
		if price <= 0 {
			price = p.OrginalCost
		}
		items = append(items, RawItem{
			ExternalID: fmt.Sprintf("%d", p.ID),
			NameFa:     strings.TrimSpace(p.FaName),
			BrandName:  strings.TrimSpace(p.BrandEnName),
			ImageURL:   strings.TrimSpace(p.Image),
			ProductURL: absoluteURL(origin, p.ProductURL),
			PriceToman: price,
			InStock:    p.StoreState,
			SlugHint:   slugFromPath(p.ProductURL, p.ID),
			Raw:        raw,
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("darukade listing produced no usable items")
	}
	return items, nil
}

func nextData(body []byte) ([]byte, error) {
	s := string(body)
	const marker = `<script id="__NEXT_DATA__" type="application/json">`
	if i := strings.Index(s, marker); i >= 0 {
		rest := s[i+len(marker):]
		j := strings.Index(rest, "</script>")
		if j < 0 {
			return nil, fmt.Errorf("truncated __NEXT_DATA__")
		}
		return []byte(rest[:j]), nil
	}
	trim := strings.TrimSpace(s)
	if strings.HasPrefix(trim, "{") {
		return []byte(trim), nil
	}
	return nil, fmt.Errorf("page is not darukade next.js listing data")
}

func absoluteURL(origin, path string) string {
	if path == "" {
		return origin
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.TrimRight(origin, "/") + path
}

func slugFromPath(path string, id int64) string {
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	last = strings.TrimSuffix(last, fmt.Sprintf("-%d", id))
	return textfa.Slugify(last)
}
