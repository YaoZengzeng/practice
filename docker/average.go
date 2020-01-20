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
	done := make(chan struct{})
	now := time.Now()
	var wgc, wgr sync.WaitGroup
	wgc.Add(times)
	wgr.Add(times)
	for i := 0; i < times; i++ {
		go func() {
			defer wgr.Done()
			var id string
			func() {
				defer wgc.Done()
				cmd := exec.Command("docker", "run", "-d", "--network", "none", image)
				out, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("run container failed: %v, output is %v\n", err, string(out))
					return
				}
				id = string(out)
			}()
			<-done
			cmd := exec.Command("docker", "rm", "-f", strings.TrimSuffix(id, "\n"))
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("remove container failed: %v, output is %v\n", err, string(out))
				return
			}
		}()
	}

	wgc.Wait()

	d := time.Since(now)
	fmt.Printf("totally %v used, %v containers/s\n", d.Seconds(), float64(times)/d.Seconds())

	close(done)

	wgr.Wait()
}
