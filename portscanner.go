//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func portScanner(host string, port string, timeout time.Duration) {
	address := host + ":" + port
	fmt.Println(address)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err == nil {
		fmt.Printf("Success: Port %s is open on %s\n", port, host)
		conn.Close()
	} else {
		if strings.Contains(err.Error(), "too many open files") {
			fmt.Println("Server might be busy. Might have to handle it later")
			return
		} else {
			fmt.Printf("Failed: Port %s is closed on %s\n", port, host)
		}
	}

}

func main() {
	// Google DNS
	portScanner("8.8.8.8", "53", 2*time.Second)

}
