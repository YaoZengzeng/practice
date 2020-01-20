package main

import (
	"flag"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	image string
	times int
)

func init() {
	flag.StringVar(&image, "image", "nginx:1.0.0", "the image of containers to create")
	flag.IntVar(&times, "times", 20, "number of times to run container")
	flag.Parse()
}

func main() {
	done, ch := make(chan struct{}), make(chan string)
	var wg sync.WaitGroup
	wg.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			defer wg.Done()
			cmd := exec.Command("docker", "run", "-d", "--network", "none", image)
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("run container failed: %v, output is %v\n", err, string(out))
				return
			}
			id := strings.TrimSuffix(string(out), "\n")

			ch <- id

			<-done
			cmd = exec.Command("docker", "rm", "-f", id)
			out, err = cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("remove container failed: %v, output is %v\n", err, string(out))
				return
			}
		}()
	}

	cnt, seconds := 0, 0
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ch:
			cnt++
			if cnt == times {
				fmt.Printf("all containers has been started\n")
				goto Finish
			}
		case <-ticker.C:
			seconds++
			fmt.Printf("seconds %v: %v containers has started\n", seconds, cnt)
		}
	}
Finish:

	close(done)

	wg.Wait()

}
