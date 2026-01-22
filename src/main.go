package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

func main() {
	// Start a new TCP listener on port 6980
	listener, err := net.Listen("tcp", ":6980")
	if err != nil {
		log.Fatal("Error listening:", err)
	}

	defer listener.Close()

	for {
		// Accept blocks until a new connection is made
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting conn:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Use buffered reader to read data from the connection
	defer conn.Close()

	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Read error: %v", err)
		return
	}

	// Using another buffer to process and write a response back
	ackMsg := strings.ToUpper(strings.TrimSpace(message))
	response := fmt.Sprintf("ACK: %s\n", ackMsg)
	// Write functions returns the number of bytes written and an error if occurred
	_, err = conn.Write([]byte(response))
	if err != nil {
		log.Printf("Server write error: %v", err)
	}
}
