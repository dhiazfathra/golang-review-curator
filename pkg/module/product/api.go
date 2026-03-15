package product

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/api/v1")
	g.POST("/products", h.register)
	g.GET("/products", h.list)
}

func (h *Handler) register(c echo.Context) error {
	var req struct {
		Name       string `json:"name"`
		Platform   string `json:"platform"`
		ProductURL string `json:"product_url"`
		ProductID  string `json:"product_id"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p, err := h.service.Register(c.Request().Context(), req.Name, req.Platform, req.ProductURL, req.ProductID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *Handler) list(c echo.Context) error {
	products, err := h.service.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]any{"data": products})
}
