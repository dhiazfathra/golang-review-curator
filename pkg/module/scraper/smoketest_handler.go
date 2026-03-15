package scraper

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const TaskSmokeRunAll = "smoke:run_all"

var smokeFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "selector_smoke_failure_total",
	Help: "Total smoke test failures by platform.",
}, []string{"platform"})

type SmokeHandler struct {
	smoke *SmokeTest
}

func NewSmokeHandler(smoke *SmokeTest) *SmokeHandler {
	return &SmokeHandler{smoke: smoke}
}

func (h *SmokeHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	if err := h.smoke.RunAll(ctx); err != nil {
		smokeFailuresTotal.WithLabelValues("all").Inc()
		return fmt.Errorf("smoke handler: %w", err)
	}
	return nil
}
