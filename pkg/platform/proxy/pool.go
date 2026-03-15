package proxy

import (
	"strings"
)

func LoadFromConfig(proxyURLs []string) []*proxySlot {
	slots := make([]*proxySlot, 0, len(proxyURLs))
	for _, u := range proxyURLs {
		u = strings.TrimSpace(u)
		if u != "" {
			slots = append(slots, &proxySlot{URL: u, healthScore: 10})
		}
	}
	return slots
}
