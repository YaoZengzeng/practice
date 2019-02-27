package context_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func BenchmarkWithTimeout(b *testing.B) {
	for concurrency := 40; concurrency <= 4e5; concurrency *= 100 {
		name := fmt.Sprintf("concurrency#%v", concurrency)
		b.Run(name, func(b *testing.B) {
			benchmarkWithTimeout(b, concurrency)
		})
	}
}

func benchmarkWithTimeout(b *testing.B, parallelContext int) {
	gomaxprocs := runtime.GOMAXPROCS(0)
	perPContext := parallelContext / gomaxprocs
	cc := make([][]context.CancelFunc, gomaxprocs)
	root := context.Background()

	wg := sync.WaitGroup{}
	for n := range cc {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c := make([]context.CancelFunc, perPContext)
			for i := range c {
				_, c[i] = context.WithTimeout(root, time.Hour)
			}
			cc[n] = c
		}(n)
	}
	wg.Wait()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		c := make([]context.CancelFunc, 10)
		for pb.Next() {
			for i := range c {
				_, c[i] = context.WithTimeout(root, time.Hour)
			}
			for _, f := range c {
				f()
			}
		}
	})
	b.StopTimer()

	for _, c := range cc {
		for _, f := range c {
			f()
		}
	}
}

func BenchmarkCancelTree(b *testing.B) {
	depths := []int{1, 10, 100, 1000}
	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth#%v", depth), func(b *testing.B) {
			b.Run(fmt.Sprintf("Root=Background"), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					root := context.Background()
					buildContextTree(root, depth)
				}
			})
			b.Run(fmt.Sprintf("Root=OpenCancel"), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					root, cancel := context.WithCancel(context.Background())
					buildContextTree(root, depth)
					cancel()
				}
			})
			b.Run(fmt.Sprintf("Root=CloseCancel"), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					root, cancel := context.WithCancel(context.Background())
					cancel()
					buildContextTree(root, depth)
				}
			})
		})
	}
}

func buildContextTree(root context.Context, depth int) {
	for i := 0; i < depth; i++ {
		root, _ = context.WithCancel(root)
	}
}
