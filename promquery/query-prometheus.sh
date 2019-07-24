#!/bin/sh

curl -G "http://172.16.0.45:9090/api/v1/query_range?" --data-urlencode 'query=(go_goroutines{kubernetes_namespace="cce-monitor", kubernetes_pod="copaddon-prometheus-server-0"} > 200)' --data-urlencode 'start=1563852148' --data-urlencode 'end=1563852328' --data-urlencode 'step=10'

