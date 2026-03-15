package scraper

import (
	"crypto/sha256"
	"fmt"
)

func dedupeHash(r CrawlResult) string {
	payload := r.RawJSON
	if len(payload) > 100 {
		payload = payload[:100]
	}
	sum := sha256.Sum256([]byte(string(r.Platform) + r.ProductURL + string(payload)))
	return fmt.Sprintf("%x", sum)
}
