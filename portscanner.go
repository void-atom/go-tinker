//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net"
	"time"
)

func portScanner(host string, port string, timeout time.Duration) {
	address := host + ":" + port
	fmt.Println(address)

	// Simple TCP connection to test if the port is listening
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		fmt.Printf("Port %s is not available: %v\n", port, err)
		return
	}
	defer conn.Close()

	fmt.Printf("Port %s is open and available!\n", port)
}
func main() {
	// Google DNS
	portScanner("8.8.8.8", "53", 2*time.Second)

}
