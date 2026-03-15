package selector

import (
	"strconv"

	"github.com/go-rod/rod"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var extractionByRule = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "scraper_extraction_rule_hits_total",
	Help: "Extraction successes by platform, field, and fallback rule index.",
}, []string{"platform", "field", "rule_index"})

var extractionFailures = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "scraper_extraction_failures_total",
	Help: "Extraction failures after all fallback rules exhausted.",
}, []string{"platform", "field"})

type Extractor interface {
	Element(selector string) (string, bool)
	XPathElement(xpath string) (string, bool)
}

func ExtractField(e Extractor, cfg SelectorConfig) (string, bool) {
	for i, rule := range cfg.Rules {
		var val string
		var ok bool
		switch rule.Type {
		case RuleTypeCSS:
			val, ok = e.Element(rule.Value)
		case RuleTypeXPath:
			val, ok = e.XPathElement(rule.Value)
		}
		if ok && val != "" {
			extractionByRule.WithLabelValues(cfg.Platform, cfg.Field, strconv.Itoa(i)).Inc()
			return val, true
		}
	}
	extractionFailures.WithLabelValues(cfg.Platform, cfg.Field).Inc()
	return "", false
}

type RodExtractor struct{ Page *rod.Page }

func (r *RodExtractor) Element(css string) (string, bool) {
	el, err := r.Page.Element(css)
	if err != nil {
		return "", false
	}
	text, err := el.Text()
	if err != nil || text == "" {
		return "", false
	}
	return text, true
}

func (r *RodExtractor) XPathElement(xpath string) (string, bool) {
	el, err := r.Page.ElementX(xpath)
	if err != nil {
		return "", false
	}
	text, err := el.Text()
	if err != nil || text == "" {
		return "", false
	}
	return text, true
}
