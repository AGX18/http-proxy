package main

import (
	"fmt"

	"github.com/AGX18/http-proxy/utils"
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
		handleClientConnection(client_sock)
		
		unix.Close(client_sock)
	}
 }

 func handleClientConnection(client_sock int) error{
	for {

		parser := utils.NewParser()
		data := make([]byte, 4096)
		upstream_sock, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
		if err != nil {
			fmt.Println("Error creating upstream socket:", err)
			return err
		}
		upstream_addr := &unix.SockaddrInet4{Port: 9090, Addr: [4]byte{127, 0, 0, 1}}
		// connect to the upstream server
		err = unix.Connect(upstream_sock, upstream_addr)
		
		// parse the data until we have a full request
		for parser.State != utils.StateDone {
			data_size, err := unix.Read(client_sock, data)
			if err != nil {
				fmt.Println("Error reading from client:", err)
			}
			if data_size == 0 {
				fmt.Println("No more data from client")
				return nil
			}
			// parse the data in the http parser
			parser.Parse(data[:data_size])
			// forward the data to the upstream server
			unix.Write(upstream_sock, data[:data_size])

			fmt.Printf("Received data from client sized (%d)\n", data_size)
			
		}
		if err == unix.ECONNREFUSED {
			fmt.Println("Bad gateway: upstream server is down")
			unix.Write(client_sock, []byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
			unix.Close(upstream_sock)
			unix.Close(client_sock)
			return err
			} else if err != nil {
				fmt.Println("Error connecting to upstream server:", err)
			unix.Close(upstream_sock)
			return err
		}
		fmt.Printf("Connected to upstream server at: %d.%d.%d.%d\n", upstream_addr.Addr[0], upstream_addr.Addr[1], upstream_addr.Addr[2], upstream_addr.Addr[3])

		res := make([]byte, 4096)


		// read the response from the upstream server
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

		if shouldCloseConnection(*parser.Request) {
			fmt.Println("Closing connection as per Connection header")
			unix.Close(upstream_sock)
			return nil
		}
	}
 }

 func shouldCloseConnection(req utils.Request) bool {
	connHeader, ok := req.Headers["Connection"]; 
	if req.Version == "HTTP/1.1" && (!ok || (ok && connHeader != "close")) {
		return false // Keep-alive by default in HTTP/1.1
	}
	return !(req.Version == "HTTP/1.0" && (ok && connHeader == "keep-alive"))
}