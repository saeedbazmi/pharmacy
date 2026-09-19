package ingest

import (
	"fmt"
	"strings"

	"github.com/saeedbazmi/pharmacy/backend/internal/platform/textfa"
)

func productSlug(item RawItem) string {
	base := item.SlugHint
	if base == "" {
		base = textfa.Slugify(item.NameEn)
	}
	if base == "" {
		base = textfa.Slugify(item.BrandName + "-" + item.ExternalID)
	}
	if base == "" {
		base = "item-" + item.ExternalID
	}
	if item.ExternalID != "" && !strings.HasSuffix(base, "-"+item.ExternalID) {
		base = base + "-" + item.ExternalID
	}
	return base
}

func uniqueSlug(base string, exists func(string) bool) string {
	if !exists(base) {
		return base
	}
	for i := 2; i < 50; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return fmt.Sprintf("%s-%d", base, 99)
}

func brandSlug(name string) string {
	s := textfa.Slugify(name)
	if s == "" {
		return "unknown-brand"
	}
	return s
}
