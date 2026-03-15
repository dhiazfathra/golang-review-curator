package scraper

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"review-curator/pkg/platform/database"
	"review-curator/pkg/platform/selector"
)

type APIHandler struct {
	crawl     *CrawlService
	selectors *selector.SelectorStore
}

func NewAPIHandler(crawl *CrawlService, selectors *selector.SelectorStore) *APIHandler {
	return &APIHandler{crawl: crawl, selectors: selectors}
}

func (h *APIHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1")

	g.POST("/crawl/jobs", h.enqueueJob)
	g.GET("/crawl/jobs/:id", h.getJob)
	g.GET("/crawl/jobs", h.listJobs)
	g.POST("/crawl/jobs/:id/retry", h.retryJob)

	g.GET("/selectors", h.listSelectors)
	g.PUT("/selectors/:id", h.updateSelector)
	g.GET("/selectors/health", h.selectorHealth)
}

func (h *APIHandler) enqueueJob(c echo.Context) error {
	var req struct {
		Platform   string `json:"platform"`
		ProductURL string `json:"product_url"`
		ProductID  string `json:"product_id"`
		MaxPages   int    `json:"max_pages"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.MaxPages == 0 {
		req.MaxPages = 10
	}
	job, err := h.crawl.EnqueueJob(
		c.Request().Context(),
		Platform(req.Platform),
		req.ProductURL,
		req.ProductID,
		req.MaxPages,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, job)
}

func (h *APIHandler) getJob(c echo.Context) error {
	job, err := h.crawl.GetJob(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, job)
}

func (h *APIHandler) listJobs(c echo.Context) error {
	platform := c.QueryParam("platform")
	status := c.QueryParam("status")
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit == 0 {
		limit = 20
	}
	jobs, total, err := h.crawl.ListJobs(c.Request().Context(), platform, status, database.Page{
		Limit: limit, Offset: offset, SortBy: "enqueued_at", SortDir: "desc",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data":       jobs,
		"pagination": map[string]any{"total": total, "limit": limit, "offset": offset},
	})
}

func (h *APIHandler) retryJob(c echo.Context) error {
	job, err := h.crawl.GetJob(c.Request().Context(), c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	if job.Status != "failed" {
		return echo.NewHTTPError(http.StatusBadRequest, "only failed jobs can be retried")
	}
	_, err = h.crawl.EnqueueJob(
		c.Request().Context(),
		job.Platform, job.ProductURL, job.ProductID, job.MaxPages,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "requeued"})
}

func (h *APIHandler) listSelectors(c echo.Context) error {
	all := h.selectors.All()
	return c.JSON(http.StatusOK, map[string]any{"data": all})
}

func (h *APIHandler) updateSelector(c echo.Context) error {
	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued for reload"})
}

func (h *APIHandler) selectorHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "see /metrics for scraper_extraction_failures_total"})
}
