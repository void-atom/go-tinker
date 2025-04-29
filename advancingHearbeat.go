// go:build ignore
package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

const defaultPingInterval = 30 * time.Second

// Pinger function that sends "ping" on the connection at intervals, can be reset
func Pinger(ctx context.Context, w net.Conn, reset <-chan time.Duration) {
	var interval time.Duration
	select {
	case <-ctx.Done():
		return
	case interval = <-reset:
	default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case newInterval := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}
		case <-timer.C:
			if _, err := w.Write([]byte("ping")); err != nil {
				return
			}
		}
		_ = timer.Reset(interval)
	}
}

func main() {
	done := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	begin := time.Now()

	// Server goroutine
	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			return
		}
		defer conn.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		resetTimer := make(chan time.Duration, 1)
		resetTimer <- time.Second
		go Pinger(ctx, conn, resetTimer)

		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			fmt.Println("SetDeadline error:", err)
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				return
			}
			fmt.Printf("[Server %s] %s\n", time.Since(begin).Truncate(time.Second), buf[:n])
			resetTimer <- 0
			err = conn.SetDeadline(time.Now().Add(5 * time.Second))
			if err != nil {
				fmt.Println("SetDeadline error:", err)
				return
			}
		}
	}()

	// Client logic
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for i := 0; i < 20; i++ {
		n, err := conn.Read(buf)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[Client %s] %s\n", time.Since(begin).Truncate(time.Second), buf[:n])

		if (i+1)%4 == 0 {
			_, err = conn.Write([]byte("PONG!!!"))
		}

	}

	<-done
	end := time.Since(begin).Truncate(time.Second)
	fmt.Printf("[Done at %s]\n", end)
}
