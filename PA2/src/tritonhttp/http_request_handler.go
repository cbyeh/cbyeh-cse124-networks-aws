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

	// Start a loop for reading requests continuously
	// Set a timeout for read operation
	// Read from the connection socket into a buffer
	// Validate the request lines that were read
	// Handle any complete requests
	// Update any ongoing requests
	// If reusing read buffer, truncate it before next read

	const timeout = 5 * time.Second
	const delim = "\r\n"
	remaining := ""

	var NewHttpRequestHeader HttpRequestHeader
	NewHttpRequestHeader.InitialLine = ""
	NewHttpRequestHeader.Host = ""
	NewHttpRequestHeader.Connection = ""
	defer conn.Close()
	defer log.Println("Closed connection")
	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		buf := make([]byte, 10)
		size, err := conn.Read(buf)
		if size > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
		}

		if err != nil {
			println("done")
			break
		}

		data := buf[:size]
		remaining = remaining + string(data)

		for strings.Contains(remaining, delim) {
			idx := strings.Index(remaining, delim)

			line := remaining[:idx]

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
