// go:build ignore

package main

import (
	"context"
	"fmt"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

// Pinger function writes ping messages at regular intervals
func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	// Set initial interval, either from reset channel or default
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}

	if interval <= 0 {
		interval = defaultPingInterval
	}

	// Create a timer with the selected interval
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	// Start the infinite loop to handle the context, reset, and timer expiry
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			// Stop previous timer and reset with new interval
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C:
			// On timer expiry, write a ping message
			if _, err := w.Write([]byte("ping")); err != nil {
				// Handle write error (e.g., if the writer is closed)
				return
			}
		}

		// Reset the timer with the current interval
		_ = timer.Reset(interval)
	}
}

// ExamplePinger demonstrates using the Pinger function with io.Pipe
func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// io.Pipe simulates a network connection (reader and writer)
	r, w := io.Pipe()

	// Create a done channel to ensure goroutines exit properly
	done := make(chan struct{})

	// Reset timer channel with initial interval
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second // initial ping interval

	// Run the Pinger function in a goroutine
	go func() {
		Pinger(ctx, w, resetTimer)
		close(done)
	}()

	// Function to simulate receiving a ping and resetting the timer
	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d
		}
		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	// Simulate various ping intervals
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	// Cancel context and wait for Pinger to finish
	cancel()
	<-done
}

func main() {
	// This is from Networking in Go book
	ExamplePinger()
}
