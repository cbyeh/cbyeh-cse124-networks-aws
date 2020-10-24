package tritonhttp

import (
	"log"
	"net"
	"strings"
	"time"
)

/*
For a connection, keep handling requests until
	1. a timeout occurs or
	2. client closes connection or
	3. client sends a bad request
*/
func (hs *HttpServer) handleConnection(conn net.Conn) {

	defer conn.Close()

	buf := make([]byte, 10)
	remaining := ""
	isFirstLine := true

	var NewHttpRequestHeader HttpRequestHeader

	// Start a loop for reading requests continuously
	for {

		// Set a timeout for read operation
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read from the connection socket into a buffer
		size, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
			conn.Close()
			return
		}
		data := buf[:size]
		remaining += string(data)

		// Validate the request lines that were read
		for strings.Contains(remaining, "\r\n") {
			idx := strings.Index(remaining, "\r\n")
			line := remaining[:idx]

			// Update any ongoing requests
			remaining = remaining[idx+1:]

			// Done with request
			if line == "\n" {
				// Handle any complete requests
				if NewHttpRequestHeader.Host != "" && NewHttpRequestHeader.InitialLine != "" {
					hs.handleResponse(&NewHttpRequestHeader, conn)
				} else {
					hs.handleBadRequest(conn)
				}
				isFirstLine = true
				return
			}

			// Fill request header, check for errors
			if isFirstLine == true { // Special case for initial line
				isFirstLine = false
				tokens := strings.Split(line, " ")
				if tokens[1][:1] == "/" && tokens[2] == "HTTP/1.1" {
					NewHttpRequestHeader.InitialLine = line
				} else {
					hs.handleBadRequest(conn)
					return
				}
			} else {
				split := strings.Split(line, " ")
				key := split[0]
				value := split[1]
				if strings.Contains(key, "Host:") {
					NewHttpRequestHeader.Host = value
				} else if strings.Contains(key, "Connection") {
					NewHttpRequestHeader.Connection = value
				}
			}

		}
	}

}
