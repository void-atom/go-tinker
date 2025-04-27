// go:build ignore
package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	// Create a listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on", listener.Addr())

	done := make(chan bool)

	// Start a goroutine to accept connections
	go func() {
		defer func() { done <- true }()

		for {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}

			// Start a goroutine to handle each connection
			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- true
				}()

				buf := make([]byte, 8)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							fmt.Println(err)
						}
						return
					}
					fmt.Printf("received: %q\n", buf[:n])
				}
			}(conn)
		}
	}()

	// Dial the listener
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send some data!
	// conn.Write([]byte("hello from client"))
	conn.Close()

	// Wait for the handler goroutine to finish
	<-done

	listener.Close()
	<-done
}
