package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ScraperJobsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scraper_jobs_total",
		Help: "Total crawl jobs by platform and status.",
	}, []string{"platform", "status"})

	ScraperCaptchaEncounters = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scraper_captcha_encounters_total",
		Help: "Captcha encounters by platform and type.",
	}, []string{"platform", "type"})

	ProxyPoolSlots = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "proxy_pool_available_slots",
		Help: "Number of healthy proxy slots currently available.",
	})

	NormaliserReviewsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "normaliser_reviews_total",
		Help: "Total normalised reviews by platform.",
	}, []string{"platform"})
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
