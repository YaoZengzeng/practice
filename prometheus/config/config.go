package config

import (
	"flag"
)

var PrometheusURL string

func init() {
	flag.StringVar(&PrometheusURL, "prom-url", "http://10.32.0.2:9090", "Prometheues URL")
}
