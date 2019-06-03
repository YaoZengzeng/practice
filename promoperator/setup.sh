#!/bin/sh

kubectl apply -f bundle.yaml

kubectl create secret generic additional-configs --from-file=prometheus-additional.yaml --dry-run -o yaml | kubectl apply -f -

kubectl apply -f prometheus.yaml

kubectl apply -f rules-deadman.yaml

kubectl apply -f rules-targetdown.yaml

kubectl create secret generic alertmanager-example --from-file=alertmanager.yaml --dry-run -o yaml | kubectl apply -f -

kubectl apply -f alertmanager-example.yaml
