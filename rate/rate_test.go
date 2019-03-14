package rate

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	if Limit(10) == Inf {
		t.Errorf("Limit(10) should not be Inf")
	}
}

func closeEnough(a, b Limit) bool {
	return math.Abs(float64(a/b)-1) < 1e-9
}

func TestEvery(t *testing.T) {
	tc := []struct {
		interval time.Duration
		target   Limit
	}{
		{
			0, Inf,
		},
		{
			-1, Inf,
		},
		{
			1 * time.Nanosecond, Limit(1e9),
		},
		{
			100 * time.Nanosecond, Limit(1e7),
		},
		{
			time.Second, Limit(1),
		},
		{
			time.Duration(2.5 * float64(time.Second)), Limit(0.4),
		},
		{
			time.Duration(math.MaxInt64), Limit(1e9 / float64(math.MaxInt64)),
		},
	}

	for _, c := range tc {
		got := Every(c.interval)
		if !closeEnough(got, c.target) {
			t.Errorf("Every(%v) = %v, want %v", c.interval, got, c.target)
		}
	}
}

const d = 100 * time.Millisecond

var (
	t0 = time.Now()
	t1 = t0.Add(d)
	t2 = t0.Add(2 * d)
	t3 = t0.Add(3 * d)
	t4 = t0.Add(4 * d)
	t5 = t0.Add(5 * d)
	t6 = t0.Add(6 * d)
)

type allow struct {
	t  time.Time
	n  int
	ok bool
}

func run(t *testing.T, limiter *Limiter, allows []allow) {
	for _, a := range allows {
		got := limiter.AllowN(a.t, a.n)
		if a.ok != got {
			t.Errorf("run limiter.AllowN want %v, got %v", a.ok, got)
		}
	}
}

func TestLimiterBurst1(t *testing.T) {
	run(t, NewLimiter(10, 1), []allow{
		{t0, 1, true},
		{t0, 1, false},
		{t0, 1, false},
		{t1, 1, true},
		{t1, 1, false},
		{t1, 1, false},
		{t2, 2, false},
		{t2, 1, true},
		{t2, 1, false},
	})
}

func TestLimiterBurst3(t *testing.T) {
	run(t, NewLimiter(10, 3), []allow{
		{t0, 2, true},
		{t0, 1, true},
		{t0, 1, false},
		{t1, 1, true},
		{t1, 1, false},
		{t2, 1, true},
		{t3, 1, true},
		{t4, 0, true},
		{t5, 0, true},
		{t6, 3, true},
	})
}

func TestSimultaneousRequests(t *testing.T) {
	var (
		limit      = 1
		burst      = 5
		numRequest = 15
	)

	var wg sync.WaitGroup
	var count int32

	wg.Add(numRequest)
	limiter := NewLimiter(Limit(limit), burst)
	f := func() {
		defer wg.Done()
		if ok := limiter.Allow(); ok {
			atomic.AddInt32(&count, 1)
		}
	}

	for i := 0; i < numRequest; i++ {
		go f()
	}
	wg.Wait()

	if int32(burst) != count {
		t.Errorf("Simultaneous Request: want %v, got %v", numRequest, count)
	}
}

func TestLongRunningQPS(t *testing.T) {
	var (
		limit = 100
		burst = 100
	)

	var wg sync.WaitGroup
	var count int32

	limiter := NewLimiter(Limit(limit), burst)

	f := func() {
		if ok := limiter.Allow(); ok {
			atomic.AddInt32(&count, 1)
		}
		wg.Done()
	}

	start := time.Now()
	end := start.Add(5 * time.Second)
	for time.Now().Before(end) {
		wg.Add(1)
		go f()

		time.Sleep(10 * time.Microsecond)
	}
	wg.Wait()
	elapsed := time.Since(start)
	ideal := float64(burst) + elapsed.Seconds()*float64(limit)

	if int32(ideal+1) < count {
		t.Errorf("ideal + 1 = %v should smaller than count = %v", int32(ideal+1), count)
	}
	if int32(0.999*ideal) > count {
		t.Errorf("0.999 * ideal = %v should smaller than count = %v", int32(0.999*ideal), count)
	}
}
