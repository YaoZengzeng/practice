package context_test

import (
	"context"
	"sync"
	"testing"
	"time"
)

type otherContext struct {
	context.Context
}

func TestWithCancel(t *testing.T) {
	c1, cancel := context.WithCancel(context.Background())
	o := &otherContext{c1}
	c2, _ := context.WithCancel(c1)
	contexts := []context.Context{c1, o, c2}

	for i, c := range contexts {
		if ch := c.Done(); ch == nil {
			t.Errorf("contexts[%d].Done() should not be nil", i)
		}
		if err := c.Err(); err != nil {
			t.Errorf("contexts[%d].Err() should be nil", i)
		}

		select {
		case <-c.Done():
			t.Errorf("contexts[%d].Done() should be blocked before cancel()", i)
		default:
		}
	}

	cancel()

	// Let the cancel propagate.
	time.Sleep(100 * time.Millisecond)

	for i, c := range contexts {
		if err := c.Err(); err != context.Canceled {
			t.Errorf("context[%d].Err() should be context.Canceled after cancel()", i)
		}

		select {
		case <-c.Done():
		default:
			t.Errorf("contexts[%d].Done() should not be blocked after cancel()", i)
		}
	}
}

func TestParentFinishesChild(t *testing.T) {
	// Context tree:
	// parent -> cancelChild
	// parent -> valueChild -> timerChild
	parent, cancel := context.WithCancel(context.Background())
	cancelChild, stop := context.WithCancel(parent)
	defer stop()
	valueChild := context.WithValue(parent, "key", "value")
	timerChild, stop := context.WithTimeout(valueChild, 100*time.Hour)
	defer stop()

	cancel()

	check := func(ctx context.Context, name string) {
		select {
		case <-ctx.Done():
		default:
			t.Errorf("%v.Done() should be closed", name)
		}

		if err := ctx.Err(); err != context.Canceled {
			t.Errorf("%v.Err() should be context.Canceled", name)
		}
	}

	check(parent, "parent")
	check(cancelChild, "cancelChild")
	check(valueChild, "valueChild")
	check(timerChild, "timerChild")

	parentCancel := context.WithValue(parent, "key", "value")
	check(parentCancel, "parentCancel")
}

func TestChildFinishesFirst(t *testing.T) {
	cancelCtx, stop := context.WithCancel(context.Background())
	defer stop()

	for _, parent := range []context.Context{context.Background(), cancelCtx} {
		child, cancel := context.WithCancel(parent)

		cancel()

		select {
		case <-child.Done():
		default:
			t.Errorf("child.Done() should be closed")
		}

		if err := child.Err(); err != context.Canceled {
			t.Errorf("child.Err() should be context.Canceled")
		}

		select {
		case <-parent.Done():
			t.Errorf("parent.Done() should be blocked")
		default:
		}

		if err := parent.Err(); err != nil {
			t.Errorf("parent.Err() should be nil")
		}
	}
}

func testDeadline(t *testing.T, name string, ctx context.Context, failtime time.Duration) {
	select {
	case <-time.After(failtime):
		t.Errorf("context %v should not timeout", name)
	case <-ctx.Done():
	}

	if err := ctx.Err(); err != context.DeadlineExceeded {
		t.Errorf("%v.Err() should be context.DeadlineExceeded", name)
	}
}

func TestDeadline(t *testing.T) {
	ctx, _ := context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
	testDeadline(t, "Normal Deadline", ctx, time.Second)

	ctx, _ = context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
	o := &otherContext{ctx}
	testDeadline(t, "Deadline + otherContext", o, time.Second)

	ctx, _ = context.WithDeadline(context.Background(), time.Now().Add(50*time.Millisecond))
	o = &otherContext{ctx}
	ctx, _ = context.WithDeadline(ctx, time.Now().Add(2*time.Second))
	testDeadline(t, "Deadline + otherContext + Deadline", ctx, time.Second)

	ctx, _ = context.WithDeadline(context.Background(), time.Now().Add(-50*time.Millisecond))
	testDeadline(t, "Deadline + past", ctx, time.Millisecond)

	ctx, _ = context.WithDeadline(context.Background(), time.Now())
	testDeadline(t, "Deadline + now", ctx, time.Millisecond)
}

func TestCanceledTimeout(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	o := otherContext{ctx}
	ctx, cancel := context.WithTimeout(o, 2*time.Second)

	cancel()
	time.Sleep(100 * time.Millisecond) // Let the cancelation propagate.

	select {
	case <-ctx.Done():
	default:
		t.Errorf("ctx.Done() should be closed")
	}
	if err := ctx.Err(); err != context.Canceled {
		t.Errorf("ctx.Err() should be context.Canceled")
	}
}

func TestSimultaneousCancel(t *testing.T) {
	root, cancel := context.WithCancel(context.Background())
	m := map[context.Context]context.CancelFunc{root: cancel}
	p := []context.Context{root}

	// Create cancel tree.
	for len(m) <= 100 {
		ctx := p[0]
		p = p[1:]
		for i := 0; i < 4; i++ {
			c, cancel := context.WithCancel(ctx)
			p = append(p, c)
			m[c] = cancel
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(m))

	// Cancel the context in random.
	for _, cancel := range m {
		go func(cancel context.CancelFunc) {
			cancel()
			wg.Done()
		}(cancel)
	}

	for ctx := range m {
		select {
		case <-ctx.Done():
		case <-time.After(1 * time.Second):
			t.Errorf("ctx.Done() should be closed")
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Errorf("channel done should be closed")
	}
}

func TestWithCancelCanceledParent(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cancel()

	child, _ := context.WithCancel(parent)
	select {
	case <-child.Done():
	case <-time.After(time.Second):
		t.Errorf("timeout wait for child.Done()")
	}

	if err := child.Err(); err != context.Canceled {
		t.Errorf("child.Err() should be context.Canceled")
	}
}

type key1 int
type key2 int

var k1 = key1(1)
var k2 = key2(1)
var k3 = key2(3)

func TestValues(t *testing.T) {
	checkValue := func(ctx context.Context, name, c1, c2, c3 string) {
		if value, ok := ctx.Value(k1).(string); ok != (len(c1) != 0) || value != c1 {
			t.Errorf("%v.Value(k1).(string) ok -> %v, len(c1) != 0 -> %v, value %v, c1 %v",
				name, ok, len(c1) != 0, value, c1)
		}
		if value, ok := ctx.Value(k2).(string); ok != (len(c2) != 0) || value != c2 {
			t.Errorf("%v.Value(k2).(string) ok -> %v, len(c2) != 0 -> %v, value %v, c2 %v",
				name, ok, len(c2) != 0, value, c2)
		}
		if value, ok := ctx.Value(k3).(string); ok != (len(c3) != 0) || value != c3 {
			t.Errorf("%v.Value(k3).(string) ok -> %v, len(c3) != 0 -> %v, value %v, c3 %v",
				name, ok, len(c3) != 0, value, c3)
		}
	}

	c0 := context.Background()
	checkValue(c0, "c0", "", "", "")

	c1 := context.WithValue(c0, k1, "c1k1")
	checkValue(c1, "c1", "c1k1", "", "")

	c2 := context.WithValue(c1, k2, "c2k2")
	checkValue(c2, "c2", "c1k1", "c2k2", "")

	c3 := context.WithValue(c2, k3, "c3k3")
	checkValue(c3, "c3", "c1k1", "c2k2", "c3k3")

	c4 := context.WithValue(c3, k1, nil)
	checkValue(c4, "c4", "", "c2k2", "c3k3")

	o0 := otherContext{context.Background()}
	checkValue(o0, "o0", "", "", "")

	o1 := otherContext{context.WithValue(o0, k1, "c1k1")}
	checkValue(o1, "o1", "c1k1", "", "")

	o2 := context.WithValue(o1, k2, "c2k2")
	checkValue(o2, "o2", "c1k1", "c2k2", "")

	o3 := otherContext{c4}
	checkValue(o3, "o3", "", "c2k2", "c3k3")

	o4 := context.WithValue(o3, k3, nil)
	checkValue(o4, "o4", "", "c2k2", "")
}
