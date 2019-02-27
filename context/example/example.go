package main

import (
	"context"
	"fmt"
	"time"
)

func ExampleWithCancel() {
	gen := func(ctx context.Context) <-chan int {
		dst := make(chan int)
		go func() {
			n := 0
			for {
				select {
				case <-ctx.Done():
					return
				case dst<-n:
					n++
				}
			}
		}()
		return dst
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := range gen(ctx) {
		fmt.Println(i)
		if i == 5 {
			return
		}
	}
	// Output:
	// 0
	// 1
	// 2
	// 3
	// 4
	// 5
}

func ExampleWithDeadline() {
	t := time.Now().Add(1 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), t)

	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("Overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	}
	// Output:
	// context deadline exceeded
}

func ExampleWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Millisecond)

	defer cancel()

	select {
	case <-time.After(1 * time.Second):
		fmt.Println("Overslept")
	case <-ctx.Done():
		fmt.Println(ctx.Err())
	}
	// Output:
	// context deadline exceeded
}

func ExampleWithValue() {
	type favContextKey string

	f := func(ctx context.Context, key favContextKey) {
		if v := ctx.Value(key); v != nil {
			fmt.Println(v)
			return
		}
		fmt.Println("key not found")
	}

	key := favContextKey("hello")
	// The provided key must be comparable and should not be type
	// of string or any other built-in type to avoid collisions
	// between packages using context.
	ctx := context.WithValue(context.Background(), key, "world")
	f(ctx, key)
	f(ctx, favContextKey("Hello"))
	// Output:
	// world
	// key not found
}

func main() {
	ExampleWithCancel()
	ExampleWithDeadline()
	ExampleWithTimeout()
	ExampleWithValue()
}
