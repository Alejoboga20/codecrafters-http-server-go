package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	listener, err := net.Listen("tcp", "0.0.0.0:4221")

	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		handleConnection(connection)
	}

}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	reader := bufio.NewReader(connection)
	requestLine, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}

	fmt.Println("Received request: ", requestLine)
	parts := strings.Split(requestLine, " ")

	if len(parts) < 3 {
		fmt.Println("Invalid request")
		return
	}

	requestUrl := parts[1]
	fmt.Println("Request URL: ", requestUrl)

	if requestUrl == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}

	if strings.HasPrefix(requestUrl, "/echo") {
		echoSring := strings.Split(requestUrl, "/echo/")[1]
		contentLength := len(echoSring)
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, echoSring)

		connection.Write([]byte(response))
	}

	connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}
