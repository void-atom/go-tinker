package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

func server(listenerChan chan net.Listener, done chan struct{}) {
	payload := make([]byte, 1<<24) // 16MB payload
	_, err := rand.Read(payload)
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		panic(err)
	}
	listenerChan <- listener

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()
		_, err = conn.Write(payload)
		if err != nil {
			panic(err)
		}
		// Notify that server is done after writing
		done <- struct{}{}
	}()
}

func client(listenerChan chan net.Listener, done chan struct{}) {
	listener := <-listenerChan
	conn, err := net.Dial("tcp", listener.Addr().String())
	defer conn.Close()
	if err != nil {
		panic(err)
	}

	buf := make([]byte, 1<<19) // 512 KB buffer

	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Println("End of file reached")
			break
		} else if err != nil {
			panic(err)
		}
		fmt.Printf("[read %d bytes]\n", n)
	}
	done <- struct{}{}
}

func main() {
	listenerChan := make(chan net.Listener)
	done := make(chan struct{})

	go server(listenerChan, done)
	go client(listenerChan, done)

	// Wait for both server and client to finish
	<-done
	<-done
	fmt.Println("Both server and client have completed.")
}
