package main

import (
    "net/http"
    "encoding/json"
    "fmt"
    "strings"
    "io/ioutil"
)

func main() {
    m := []string{
        "up",
        "prometheus_engine_query_duration_seconds",
        "prometheus_tsdb_reloads_total",
    }
    b, err := json.Marshal(m)
    if err != nil {
        fmt.Println("marshal failed: %v", err)
        return
    }

    s := strings.NewReader(fmt.Sprintf("metrics=%s", string(b)))
    req, err := http.NewRequest("POST", "http://127.0.0.1:8080/backend/prometheus/clusters/test/namespaces/default/pods/prometheus/metrics-records", s)
    if err != nil {
        fmt.Printf("new request failed: %v", err)
        return
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Operation", "Add")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        fmt.Printf("http Do failed: %v", err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("ReadAll failed: %v", err)
        return
    }

    fmt.Println(string(body))
}
