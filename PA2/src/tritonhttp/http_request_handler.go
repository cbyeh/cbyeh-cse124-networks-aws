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
				if line[:4] == "GET " && tokens[1][:1] == "/" && tokens[2] == "HTTP/1.1" && len(tokens) == 3 {
					NewHttpRequestHeader.InitialLine = line
				} else {
					hs.handleBadRequest(conn)
					return
				}
			} else {
				if !strings.Contains(line, ":") {
					hs.handleBadRequest(conn)
					return
				}
				// Find colon and make sure key is all letters before colon
				colonIdx := strings.Index(line, ":")
				key := line[1:colonIdx]
				if strings.Contains(key, " ") {
					hs.handleBadRequest(conn)
					return
				}

				// Get value by trimming off spaces on the left
				value := strings.TrimLeft(line[colonIdx+1:], " ")
				if key == "Host" {
					NewHttpRequestHeader.Host = value
				} else if key == "Connection" {
					NewHttpRequestHeader.Connection = value
				}
			}

			// Update any ongoing requests
			remaining = remaining[idx+1:]

		}
	}

}
