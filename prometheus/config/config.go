package config

import (
	"flag"
)

var PrometheusURL string

func init() {
	flag.StringVar(&PrometheusURL, "prom-url", "http://127.0.0.1:9090", "Prometheues URL")
}
