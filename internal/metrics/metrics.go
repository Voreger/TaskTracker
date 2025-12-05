package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

var (
	HTTPRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP запросов",
		},
		[]string{"path", "method", "status"},
	)

	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность HTTP запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	TaskCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "task_created_total",
			Help: "Общее количество созданных задач",
		})
)

func init() {
	prometheus.MustRegister(HTTPRequests, HTTPDuration, TaskCreated)
}

func StatusToString(code int) string {
	return strconv.Itoa(code)
}

func ObserveDuration(path, method string, d time.Duration) {
	HTTPDuration.WithLabelValues(path, method).Observe(d.Seconds())
}

func IncRequest(path, method string, status int) {
	HTTPRequests.WithLabelValues(path, method, StatusToString(status)).Inc()
}
