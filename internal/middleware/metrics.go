package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	URLsCreated         prometheus.Counter
}

func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		URLsCreated: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "urls_created_total",
			Help:      "Total number of shortened URLs",
		}),
	}
}

func PrometheusMiddleware(metrics *Metrics) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			path := c.Path()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			status := strconv.Itoa(c.Response().Status)
			duration := time.Since(start).Seconds()

			metrics.HTTPRequestsTotal.WithLabelValues(
				c.Request().Method,
				path,
				status,
			).Inc()

			metrics.HTTPRequestDuration.WithLabelValues(
				c.Request().Method,
				path,
			).Observe(duration)

			return nil
		}
	}
}
