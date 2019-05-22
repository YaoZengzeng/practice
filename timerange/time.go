package main

import (
	"time"
	"fmt"
)

func main() {
	now := time.Now()
	end := now.Unix()
	start := now.Add(-5 * time.Minute).Unix()

	fmt.Printf("start: %v\nend: %v\n", start, end)
}
