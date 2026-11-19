package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/vitalii-q/selena-users-service/internal/metrics"
)

func TestPrometheusMiddlewareUsesRouteTemplateAndNormalizesMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/users/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	beforeCount := counterValue(t, "GET", "/users/:id", "204")
	beforeDurationCount := histogramCount(t, "GET", "/users/:id", "204")
	request := httptest.NewRequest(http.MethodGet, "/users/user-123?token=private", nil)
	router.ServeHTTP(httptest.NewRecorder(), request)

	if got := counterValue(t, "GET", "/users/:id", "204"); got != beforeCount+1 {
		t.Fatalf("route-template counter = %v, want %v", got, beforeCount+1)
	}
	if got := histogramCount(t, "GET", "/users/:id", "204"); got != beforeDurationCount+1 {
		t.Fatalf("route-template histogram sample count = %v, want %v", got, beforeDurationCount+1)
	}
	if got := counterValue(t, "GET", "/users/user-123", "204"); got != 0 {
		t.Fatalf("raw URL path created a metric series with value %v", got)
	}
}

func TestPrometheusMiddlewareBoundsUnknownRoutesAndMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PrometheusMiddleware())

	before := counterValue(t, "OTHER", "unmatched", "404")
	for _, method := range []string{"CUSTOM-ONE", "CUSTOM-TWO"} {
		request := httptest.NewRequest(method, "/users/private-id?email=private", nil)
		router.ServeHTTP(httptest.NewRecorder(), request)
	}

	if got := counterValue(t, "OTHER", "unmatched", "404"); got != before+2 {
		t.Fatalf("normalized unmatched counter = %v, want %v", got, before+2)
	}
	if got := counterValue(t, "CUSTOM-ONE", "unmatched", "404"); got != 0 {
		t.Fatalf("unrecognized method created its own series with value %v", got)
	}
	if got := counterValue(t, "OTHER", "/users/private-id", "404"); got != 0 {
		t.Fatalf("unmatched raw URL created a series with value %v", got)
	}
}

func TestPrometheusMiddlewareSkipsMetricsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PrometheusMiddleware())
	router.GET("/metrics", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	beforeCount := counterValue(t, "GET", "/metrics", "200")
	beforeDurationCount := histogramCount(t, "GET", "/metrics", "200")
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if got := counterValue(t, "GET", "/metrics", "200"); got != beforeCount {
		t.Fatalf("metrics endpoint counter = %v, want unchanged value %v", got, beforeCount)
	}
	if got := histogramCount(t, "GET", "/metrics", "200"); got != beforeDurationCount {
		t.Fatalf("metrics endpoint histogram sample count = %v, want unchanged value %v", got, beforeDurationCount)
	}
}

func counterValue(t *testing.T, method, route, status string) float64 {
	t.Helper()
	metric := &dto.Metric{}
	if err := metrics.HTTPRequestsTotal.WithLabelValues(method, route, status).Write(metric); err != nil {
		t.Fatalf("write counter metric: %v", err)
	}
	return metric.GetCounter().GetValue()
}

func histogramCount(t *testing.T, method, route, status string) uint64 {
	t.Helper()
	metric := &dto.Metric{}
	observer := metrics.HTTPRequestDuration.WithLabelValues(method, route, status)
	collector, ok := observer.(prometheus.Metric)
	if !ok {
		t.Fatal("histogram observer does not expose a Prometheus metric")
	}
	if err := collector.Write(metric); err != nil {
		t.Fatalf("write histogram metric: %v", err)
	}
	return metric.GetHistogram().GetSampleCount()
}
