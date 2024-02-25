package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

const (
	connectionClosed int = iota
	connectionOpened
)

type event struct {
	eventType int
	conn      net.Conn
}

const port = "8080"

func main() {

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port", port)

	messages := make(chan string)
	connectionEvents := make(chan event)
	connections := map[net.Conn]struct{}{}

	go func() {
		for {
			select {
			case message := <-messages:
				{
					fmt.Printf("Message received from channel %s", message)
					for conn := range connections {
						conn.Write([]byte(message))
					}
				}
			case connectionEvent := <-connectionEvents:
				{
					switch connectionEvent.eventType {
					case connectionOpened:
						fmt.Printf("New connection: %s\n", connectionEvent.conn.RemoteAddr().String())
						connections[connectionEvent.conn] = struct{}{}
					case connectionClosed:
						fmt.Printf("Connection closed: %s\n", connectionEvent.conn.RemoteAddr().String())
						delete(connections, connectionEvent.conn)
					}
				}
			}

		}
	}()

	// Accept connections in a loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			return
		}
		// Handle connections concurrently
		go handleConnection(conn, messages, connectionEvents)
	}
}

func handleConnection(conn net.Conn, messages chan string, connectionEvents chan event) {
	defer conn.Close()
	connectionEvents <- event{connectionOpened, conn}

	fmt.Printf("Connection established from %s\n", conn.RemoteAddr())

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
	for {
		// Read data from the connection into the buffer
		length, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client:", conn.RemoteAddr())
				connectionEvents <- event{connectionClosed, conn}
				return
			}
			// print message if timeout
			if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Println("Connection timed out:", conn.RemoteAddr())
				connectionEvents <- event{connectionClosed, conn}
				return
			}
			fmt.Println("Error reading:", err.Error())
			connectionEvents <- event{connectionClosed, conn}
			return
		}

		// Process the data read from the connection
		processData(buf[:length], messages)
	}
}

func processData(data []byte, messages chan string) {
	messages <- string(data)
}
