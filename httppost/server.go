package main

import (
    "net/http"
    "fmt"
    "encoding/json"
)

func f(w http.ResponseWriter, req *http.Request) {
    b := req.PostFormValue("metrics")
    var metrics []string
    err := json.Unmarshal([]byte(b), &metrics)
    if err != nil {
        fmt.Printf("json.Unmarshal failed: %v", err)
        return
    }
    fmt.Printf("metrics: %v\n", metrics)
}

func main() {
    http.HandleFunc("/", f)
    http.ListenAndServe(":8080", nil)
}
