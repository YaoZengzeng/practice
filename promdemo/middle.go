package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	http_request_total = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "The total number of processed http requests",
		},
		[]string{"path"},
	)
	http_request_in_flight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_request_in_flight",
			Help: "Current number of http requests in flight",
		},
		[]string{"path"},
	)
	http_request_duration_seconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Histogram of lantencies for HTTP requests",
			// Buckets:	[]float64{.1, .2, .4, 1, 3, 8, 20, 60, 120},
		},
		[]string{"path"},
	)
	http_request_summary_seconds = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_summary_seconds",
			Help: "Summary of lantencies for HTTP requests",
			// Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999, 0.0001},
		},
		[]string{"path"},
	)
)

func main() {
	http.HandleFunc("/", func(http.ResponseWriter, *http.Request) {
		now := time.Now()

		http_request_in_flight.WithLabelValues("root").Inc()
		defer http_request_in_flight.WithLabelValues("root").Dec()
		http_request_total.WithLabelValues("root").Inc()

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		http_request_duration_seconds.WithLabelValues("root").Observe(time.Since(now).Seconds())
		http_request_summary_seconds.WithLabelValues("root").Observe(time.Since(now).Seconds())
	})

	http.HandleFunc("/foo", func(http.ResponseWriter, *http.Request) {
		now := time.Now()

		http_request_in_flight.WithLabelValues("foo").Inc()
		defer http_request_in_flight.WithLabelValues("foo").Dec()
		http_request_total.WithLabelValues("foo").Inc()

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		http_request_duration_seconds.WithLabelValues("foo").Observe(time.Since(now).Seconds())
		http_request_summary_seconds.WithLabelValues("foo").Observe(time.Since(now).Seconds())
	})

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}
