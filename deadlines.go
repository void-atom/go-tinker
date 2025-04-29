// go:build ignore

package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	sync := make(chan struct{})

	// Start a TCP listener on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	// Start server-side logic in a goroutine
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			return
		}
		defer func() {
			conn.Close()
			close(sync) // allow main goroutine to proceed
		}()

		// Set a 5-second deadline
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			fmt.Println("SetDeadline error:", err)
			return
		}

		// Attempt to read from connection (will timeout if no data)
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		nErr, ok := err.(net.Error)
		if !ok || !nErr.Timeout() {
			fmt.Printf("expected timeout error; actual: %v", err)
		} else {
			fmt.Println("Timeout error occured: ", err)
		}

		// Signal main goroutine to send data
		sync <- struct{}{}

		// Reset deadline and try reading real data
		err = conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			fmt.Println("Second SetDeadline error:", err)
			return
		}
		_, err = conn.Read(buf)
		if err != nil {
			fmt.Println("Second read error:", err)
		} else {
			fmt.Printf("Successfully read: %q\n", buf)
		}
	}()

	// Client connects to the listener
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Wait for server to timeout and then be ready for real data
	<-sync

	// Send data to server after timeout
	_, err = conn.Write([]byte("1"))
	if err != nil {
		panic(err)
	}

	// Read response (likely EOF after server closes)
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != io.EOF {
		fmt.Printf("Expected EOF from server close; got: %v\n", err)
	} else {
		fmt.Println("Received EOF after server close.")
	}
}
