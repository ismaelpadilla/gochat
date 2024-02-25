package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":")
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on", listener.Addr())

	con, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer con.Close()

	stdin := make(chan string)
	netin := make(chan string)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		bytes := make([]byte, 1024)
		for {
			i, err := reader.Read(bytes)
			if err != nil {
				panic(err)
			}
			stdin <- string(bytes[:i])
		}
	}()

	go handleConnection(con, netin)

	for {
		select {
		case in := <-stdin:
			{
				con.Write([]byte(in))
			}
		case in := <-netin:
			{
				fmt.Print(in)
			}
		}
	}
}

func handleConnection(conn net.Conn, netin chan string) {
	defer conn.Close()

	fmt.Printf("Connection established from %s\n", conn.RemoteAddr())

	buf := make([]byte, 1024)
	for {
		// Read data from the connection into the buffer
		conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
		length, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Connection closed by client:", conn.RemoteAddr())
				return
			}
			// print message if timeout
			if err, ok := err.(net.Error); ok && err.Timeout() {
				fmt.Println("Connection timed out:", conn.RemoteAddr())
				return
			}
			fmt.Println("Error reading:", err.Error())
			return
		}

		netin <- string(buf[:length])
	}
}
