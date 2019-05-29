package main

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Exporter struct {
	up *prometheus.Desc
}

func NewExporter() *Exporter {
	namespace := "exporter"
	up := prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "If scrape target is healthy", nil, nil)
	return &Exporter{
		up: up,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
}

func (e *Exporter) Scrape() (up float64) {
	// Scrape raw monitoring data from target, may need to do some data format conversion here
	rand.Seed(time.Now().UnixNano())
	return float64(rand.Intn(2))
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	up := e.Scrape()
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, up)
}

func main() {
	registry := prometheus.NewRegistry()

	exporter := NewExporter()

	registry.Register(exporter)

	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	http.ListenAndServe(":8080", nil)
}
