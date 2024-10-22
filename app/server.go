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
		go handleConnection(connection)
	}

}

func readRequest(connection net.Conn) (string, map[string]string, string) {
	reader := bufio.NewReader(connection)
	requestLine, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return "", nil, ""
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			fmt.Println("Error reading headers: ", err.Error())
			return "", nil, ""
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		headerPaths := strings.Split(line, ": ")
		headers[headerPaths[0]] = headerPaths[1]
	}

	var body string
	if contentLength, ok := headers["Content-Length"]; ok {
		bodyLength := 0
		fmt.Sscanf(contentLength, "%d", &bodyLength)

		// Read the body based on the Content-Length
		bodyBytes := make([]byte, bodyLength)
		_, err := reader.Read(bodyBytes)
		if err != nil {
			fmt.Println("Error reading body: ", err.Error())
			return requestLine, headers, ""
		}

		body = string(bodyBytes)
	}

	return requestLine, headers, body
}

func handleConnection(connection net.Conn) {
	defer connection.Close()
	requestLine, headers, body := readRequest(connection)

	fmt.Println("Received request: ", requestLine)
	fmt.Println("Received headers: ", headers)
	fmt.Println("Received body: ", body)

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
		handleEchoRequest(connection, requestUrl, headers)
	}
	if strings.HasPrefix(requestUrl, "/user-agent") {
		handleUserAgentRequest(connection, headers)
	}
	if strings.HasPrefix(requestUrl, "/files") {
		handleFilesRequest(connection, requestUrl, headers, body)
	}

	connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func handleEchoRequest(connection net.Conn, requestUrl string, headers map[string]string) {
	echoSring := strings.Split(requestUrl, "/echo/")[1]
	contentLength := len(echoSring)

	encodingHeader := headers["Accept-Encoding"]

	if encodingHeader == "gzip" {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\nContent-Encoding: gzip\r\n\r\n%s", contentLength, echoSring)
		connection.Write([]byte(response))
		return
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, echoSring)
	connection.Write([]byte(response))
}

func handleUserAgentRequest(connection net.Conn, headers map[string]string) {
	userAgentHeaderValue := headers["User-Agent"]
	contentLength := len(userAgentHeaderValue)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", contentLength, userAgentHeaderValue)

	connection.Write([]byte(response))
}

func handleFilesRequest(connection net.Conn, requestUrl string, headers map[string]string, body string) {
	dir := os.Args[2]

	fileName := strings.Split(requestUrl, "/files/")[1]
	filePath := dir + fileName

	if _, ok := headers["Content-Length"]; ok {
		file, err := os.Create(filePath)

		if err != nil {
			fmt.Println("Error creating file: ", err.Error())
			connection.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		defer file.Close()

		file.Write([]byte(body))
		connection.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("Error opening file: ", err.Error())
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()

	if err != nil {
		fmt.Println("Error getting file info: ", err.Error())
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	fileSize := fileInfo.Size()

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n",
		fileSize)

	connection.Write([]byte(response))

	buffer := make([]byte, 1024)
	for {
		n, err := file.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println("Error reading file: ", err.Error())
			connection.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		connection.Write(buffer[:n])
	}
}
