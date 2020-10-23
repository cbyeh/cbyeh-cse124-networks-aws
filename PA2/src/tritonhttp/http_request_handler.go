package tritonhttp

import (
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
	const timeout = 5 * time.Second
	var NewHttpRequestHeader HttpRequestHeader

	// Start a loop for reading requests continuously
	for {

		// Set a timeout for read operation
		conn.SetReadDeadline(time.Now().Add(timeout))

		// Read from the connection socket into a buffer
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]
		remaining += string(data)

		// Validate the request lines that were read
		for strings.Contains(remaining, "\n") {
			idx := strings.Index(remaining, "\n")
			line := remaining[:idx]
			tokens := strings.Split(line, " ")

			// Done with request
			if len(tokens) == 1 {
				break
			}

			// Fill request header, check for errors
			if tokens[0][len(tokens[0])-1:] != ":" {
				hs.handleBadRequest(conn)
			} else if tokens[0] == "Host:" {
				NewHttpRequestHeader.Host = tokens[1]
			} else if tokens[0] == "Connection:" {
				NewHttpRequestHeader.Connection = tokens[0]
			} else if tokens[0] == "GET" {
				NewHttpRequestHeader.InitialLine = tokens[0] + tokens[1] + tokens[2]
			}
			// Update any ongoing requests
			remaining = remaining[idx+1:]
		}
		// Handle any complete requests
		if NewHttpRequestHeader.Host != "" && NewHttpRequestHeader.Connection != "" && NewHttpRequestHeader.InitialLine != "" {
			hs.handleResponse(&NewHttpRequestHeader, conn)
		} else {
			hs.handleBadRequest(conn)
		}

	}

}
