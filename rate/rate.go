package rate

import (
	"math"
	"sync"
	"time"
)

type Limit float64

func (l Limit) tokensToDuration(tokens float64) time.Duration {
	return time.Duration(tokens / float64(l) * float64(time.Second))
}

func (l Limit) durationToTokens(d time.Duration) int {
	return int(float64(l) * d.Seconds())
}

const Inf = Limit(math.MaxFloat64)

const InfDuration = time.Duration(math.MaxInt64)

func Every(interval time.Duration) Limit {
	if interval <= 0 {
		return Inf
	}

	return 1 / Limit(interval.Seconds())
}

type Limiter struct {
	limit  Limit
	tokens int
	burst  int
	last   time.Time

	sync.Mutex
}

func NewLimiter(r Limit, b int) *Limiter {
	return &Limiter{
		limit: r,
		burst: b,
	}
}

func (lim *Limiter) Allow() bool {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now. Use this method if you intend to
// drop/skip events that exceed the rate limit. Otherwise use Reserve or Wait.
func (lim *Limiter) AllowN(now time.Time, n int) bool {
	return lim.reserveN(now, n, 0).ok
}

// ReserveN returns a Reservation that indicates how long the caller must wait before n events
// happen. The Limiter takes this Reservation into account when allowing future events. ReserveN
// returns false if n exceeds the Limiter's burst size.
func (lim *Limiter) ReserveN(now time.Time, n int) *Reservation {
	return lim.reserveN(now, n, InfDuration)
}

func (lim *Limiter) reserveN(now time.Time, n int, maxFutureTime time.Duration) *Reservation {
	lim.Lock()
	defer lim.Unlock()

	now, last, tokens := lim.advance(now, n)

	var waitDuration time.Duration
	tokens -= n
	if tokens < 0 {
		waitDuration = lim.limit.tokensToDuration(float64(-tokens))
	}

	ok := n <= lim.burst && waitDuration <= maxFutureTime

	res := &Reservation{
		ok: ok,
	}

	if ok {
		lim.last = now
		lim.tokens = tokens
	} else {
		lim.last = last
	}

	return res
}

func (lim *Limiter) advance(now time.Time, n int) (time.Time, time.Time, int) {
	last := lim.last
	if now.Before(last) {
		last = now
	}

	maxElapse := lim.limit.tokensToDuration(float64(lim.burst - lim.tokens))
	elapse := now.Sub(last)
	if elapse > maxElapse {
		elapse = maxElapse
	}

	delta := lim.limit.durationToTokens(elapse)
	tokens := lim.tokens + delta

	return now, last, tokens
}

type Reservation struct {
	ok bool
}
