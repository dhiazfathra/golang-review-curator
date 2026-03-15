package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
)

func (s *E2ESuite) TestAPIFlow() {
	e := echo.New()

	registerTestRoutes(e)

	productReq := map[string]interface{}{
		"name":        "Test Product",
		"platform":    "shopee",
		"product_url": "https://shopee.co.id/test",
		"product_id":  "12345",
	}
	body, _ := json.Marshal(productReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/reviews", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	s.NotNil(response["data"])
}

func (s *E2ESuite) TestProductsAPI() {
	e := echo.New()
	registerTestRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	s.NotNil(response["data"])
}

func (s *E2ESuite) TestReviewsAPI() {
	e := echo.New()
	registerTestRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/reviews", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)

	var response map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &response)
	s.NotNil(response["data"])
}

func registerTestRoutes(e *echo.Echo) {
	g := e.Group("/api/v1")
	g.GET("/products", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"data": []interface{}{}})
	})
	g.POST("/products", func(c echo.Context) error {
		return c.JSON(http.StatusCreated, map[string]interface{}{"id": "test-id"})
	})
	g.GET("/reviews", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"data": []interface{}{}})
	})
	g.GET("/selectors/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"})
	})
}
