package tritonhttp

import (
	"log"
	"net"
	"strings"
	"time"
	// "io"
)

/*
For a connection, keep handling requests until
	1. a timeout occurs or
	2. client closes connection or
	3. client sends a bad request
*/
func (hs *HttpServer) handleConnection(conn net.Conn) {

	const timeout = 5 * time.Second
	const delim = "\r\n"
	remaining := ""

	var NewHttpRequestHeader HttpRequestHeader
	NewHttpRequestHeader.InitialLine = ""
	NewHttpRequestHeader.Host = ""
	NewHttpRequestHeader.Connection = ""
	defer conn.Close()
	defer log.Println("Closed connection")
	// Start a loop for reading requests continuously
	for {
		// Set a timeout for read operation
		conn.SetReadDeadline(time.Now().Add(timeout))
		buf := make([]byte, 10)
		size, err := conn.Read(buf)
		if err != nil {
			break
		}
		if size > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
		}
		// Read from the connection socket into a buffer
		data := buf[:size]
		remaining = remaining + string(data)

		// Validate the request lines that were read
		for strings.Contains(remaining, delim) {
			idx := strings.Index(remaining, delim)
			line := remaining[:idx]
			// Handle any complete requests
			if NewHttpRequestHeader.InitialLine == "" {
				tokens := strings.Fields(line)
				if len(tokens) == 3 {
					if tokens[0] == "GET" && tokens[1][:1] == "/" && tokens[2] == "HTTP/1.1" {
						NewHttpRequestHeader.InitialLine = line
					}
				}
			} else {
				if strings.Contains(line, ":") {
					colonIdx := strings.Index(line, ":")
					key := line[:colonIdx]

					value := strings.Fields(line[colonIdx+1:])[0]

					if key == "Host" {
						NewHttpRequestHeader.Host = value
						println("Host: " + value)
					} else if key == "Connection" {
						NewHttpRequestHeader.Connection = value
					}
				}
			}
			// Update any ongoing requests
			remaining = remaining[idx+2:]
			if len(remaining) >= 2 && remaining[:2] == delim {
				remaining = remaining[2:]
				if NewHttpRequestHeader.Host != "" && NewHttpRequestHeader.InitialLine != "" {
					hs.handleResponse(&NewHttpRequestHeader, conn)
					NewHttpRequestHeader.InitialLine = ""
					NewHttpRequestHeader.Host = ""
					NewHttpRequestHeader.Connection = ""
				}
			} else {
				break
			}
		}
	}
}
