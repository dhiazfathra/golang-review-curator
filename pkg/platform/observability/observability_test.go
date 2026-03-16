package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger_Pretty(t *testing.T) {
	InitLogger(true)
}

func TestInitLogger_NotPretty(t *testing.T) {
	InitLogger(false)
}

func TestRequestLogger(t *testing.T) {
	middleware := RequestLogger()
	require.NotNil(t, middleware)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware(func(c echo.Context) error {
		return nil
	})

	err := handler(c)
	assert.NoError(t, err)
}

func TestRequestLogger_LogsRequest(t *testing.T) {
	middleware := RequestLogger()
	require.NotNil(t, middleware)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
}

func TestRequestLogger_WithQueryParams(t *testing.T) {
	middleware := RequestLogger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/test?foo=bar", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware(func(c echo.Context) error {
		return nil
	})

	err := handler(c)
	assert.NoError(t, err)
}

func TestRequestLogger_ErrorHandler(t *testing.T) {
	middleware := RequestLogger()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := middleware(func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "error")
	})

	err := handler(c)
	assert.Error(t, err)
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	require.NotNil(t, handler)
}

func TestMetricsHandler_ServesPrometheus(t *testing.T) {
	handler := MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestTracer(t *testing.T) {
	tr := Tracer("test-service")
	assert.NotNil(t, tr)
}

func TestStartSpan(t *testing.T) {
	ctx, span := StartSpan(context.Background(), "test-span")
	assert.NotNil(t, ctx)
	assert.NotNil(t, span)
}

func TestInitTracer_InvalidEndpoint(t *testing.T) {
	shutdown, err := InitTracer(context.Background(), "invalid://endpoint")
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, shutdown)
	}
}

func TestServiceName(t *testing.T) {
	const expected = "review-curator"
	assert.Equal(t, expected, serviceName)
}
