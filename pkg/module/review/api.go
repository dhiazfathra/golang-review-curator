package review

import (
	"net/http"
	"strconv"
	"time"

	"review-curator/pkg/platform/database"

	"github.com/labstack/echo/v4"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1")
	g.GET("/reviews", h.list)
	g.GET("/reviews/summary/:product_id", h.summary)
}

func (h *Handler) list(c echo.Context) error {
	f := ListFilter{
		Platform:  c.QueryParam("platform"),
		ProductID: c.QueryParam("product_id"),
		Language:  c.QueryParam("language"),
	}
	if r := c.QueryParam("rating"); r != "" {
		f.Rating, _ = strconv.Atoi(r)
	}
	if from := c.QueryParam("from"); from != "" {
		f.From, _ = time.Parse(time.DateOnly, from)
	}
	if to := c.QueryParam("to"); to != "" {
		f.To, _ = time.Parse(time.DateOnly, to)
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if limit == 0 {
		limit = 20
	}

	reviews, total, err := h.service.List(c.Request().Context(), f, database.Page{
		Limit: limit, Offset: offset, SortBy: "reviewed_at", SortDir: "desc",
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{
		"data":       reviews,
		"pagination": map[string]any{"total": total, "limit": limit, "offset": offset},
	})
}

func (h *Handler) summary(c echo.Context) error {
	productID := c.Param("product_id")
	platform := c.QueryParam("platform")
	summary, err := h.service.GetSummary(c.Request().Context(), platform, productID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, summary)
}
