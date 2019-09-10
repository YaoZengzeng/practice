package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	host        string
	concurrency int
)

func init() {
	flag.StringVar(&host, "host", "www.google.com", "the dns query target")
	flag.IntVar(&concurrency, "concurrency", 1000, "dns queries per second")
}

func main() {
	flag.Parse()

	t := time.NewTicker(time.Second / time.Duration(concurrency))
	defer t.Stop()

	for {
		<-t.C
		go func() {
			_, err := net.LookupHost(host)
			if err != nil {
				fmt.Fprintf(os.Stderr, "look up host failed: %v\n", err)
			}
		}()
	}
}
