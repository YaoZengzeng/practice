#!/bin/sh

kubectl apply -f prometheus-config.yml

kubectl apply -f prometheus-rbac-setup.yml

kubectl apply -f prometheus-deployment.yml

kubectl apply -f node-exporter-daemonset.yml
