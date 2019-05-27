package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name:	"http_request_total",
			Help:	"The total number of processed http requests",
		},
		[]string{"path"},
	)
)

func main() {
	http.HandleFunc("/", func(http.ResponseWriter, *http.Request){
		httpRequestTotal.WithLabelValues("root").Inc()	
	})

	http.HandleFunc("/foo", func(http.ResponseWriter, *http.Request){
		httpRequestTotal.WithLabelValues("foo").Inc()
	})

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

