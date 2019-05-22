package main

import (
    "net/http"
    "encoding/json"
    "fmt"
    "strings"
    "io/ioutil"
)

const PostURL = "http://127.0.0.1:8080/backend/prometheus/clusters/test/namespaces/default/pods/prometheus-6f5d56767b-vbvwd/series?start=1558511765&end=1558512065&step=15s"

func main() {
    m := []string{
        "up",
        "prometheus_tsdb_reloads_total",
        "prometheus_sd_discovered_targets",
    }
    b, err := json.Marshal(m)
    if err != nil {
        fmt.Println("marshal failed: %v", err)
        return
    }

    s := strings.NewReader(fmt.Sprintf("metrics=%s", string(b)))
    req, err := http.NewRequest("POST", PostURL, s)
    if err != nil {
        fmt.Printf("new request failed: %v", err)
        return
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    // req.Header.Set("Operation", "Add")

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
