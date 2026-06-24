package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// List of metrics for Prometheus monitoring
var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "users_service_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "users_service_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets, // default buckets Prometheus
		},
		[]string{"method", "path", "status"},
	)

	// Counts successfully created users.
	UsersCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "users_service_users_created_total",
			Help: "Total number of successfully created users",
		},
	)
)

func Register() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		UsersCreatedTotal,
	)
}