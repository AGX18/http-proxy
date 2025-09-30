package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// an http proxy is a server for the client and a client for the server
// it accepts connections from the client and forwards them to the server
// it then forwards the response from the server back to the client

 func main() {
	fmt.Println("Starting...")
	// we need to create a socket first  a tcp socket
	sock, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	defer unix.Close(sock)
	if err != nil {
		panic(err)
	}
	fmt.Println("Socket created:", sock)
	own_addr := &unix.SockaddrInet4{Port: 8080, Addr: [4]byte{0, 0, 0, 0}}
	err = unix.Bind(sock, own_addr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Socket bound to port 8080")

	err = unix.Listen(sock, unix.SOMAXCONN)
	if err != nil {
		panic(err)
	}

	for {

		client_sock, client_addr, err := unix.Accept(sock)
		if err != nil {
			fmt.Println("Error accepting connection:", err)
		}
		fmt.Println("Accepted connection from:", client_addr)
		// defer unix.Close(client_sock)
		
		buffer := make([]byte, 4096)
		data_size, err := unix.Read(client_sock, buffer)
		if err != nil {
			fmt.Println("Error reading from client:", err)
		}
		fmt.Printf("Received data from client sized (%d): %s\n", data_size, string(buffer[:data_size]))
		
		upstream_sock, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
		if err != nil {
			panic(err)
		}
		// defer unix.Close(upstream_sock)
		
		upstream_addr := &unix.SockaddrInet4{Port: 9090, Addr: [4]byte{127, 0, 0, 1}}
		err = unix.Connect(upstream_sock, upstream_addr)
		if err != nil {
			fmt.Println("Error connecting to upstream server:", err)
		}
		fmt.Printf("Connected to upstream server at: %d.%d.%d.%d\n", upstream_addr.Addr[0], upstream_addr.Addr[1], upstream_addr.Addr[2], upstream_addr.Addr[3])
		
		unix.Write(upstream_sock, buffer[:data_size])
		
		res := make([]byte, 4096)
		for {
			
			res_size, err := unix.Read(upstream_sock, res)
			if err != nil {
				fmt.Println("Error reading from upstream server:", err)
			}
			if res_size == 0 {
				fmt.Println("--------------------------------")
				fmt.Println("No more data from upstream server")
				break
			}
			fmt.Printf("Received data from upstream server sized (%d): %s\n", res_size, string(res[:res_size]))
			unix.Write(client_sock, res[:res_size])
		}
		unix.Close(upstream_sock)
		unix.Close(client_sock)
		
	}
 }