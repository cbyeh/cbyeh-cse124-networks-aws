package tritonhttp

import (
	"log"
	"net"
	"strings"
	"time"
	"unicode"
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
	isFirstLine := true
	remaining := ""

	var NewHttpRequestHeader HttpRequestHeader
	NewHttpRequestHeader.InitialLine = ""
	NewHttpRequestHeader.Host = ""
	NewHttpRequestHeader.Connection = ""
	NewHttpRequestHeader.IsBadRequest = false
	defer conn.Close()
	defer log.Println("Closed connection")

	// Start a loop for reading requests continuously
	for {
		// Set a timeout for read operation
		conn.SetReadDeadline(time.Now().Add(timeout))
		buf := make([]byte, 10)
		size, err := conn.Read(buf)
		if size > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
		}
		if err != nil {
			break
		}

		// Read from the connection socket into a buffer
		data := buf[:size]
		remaining = remaining + string(data)

		// Validate the request lines that were read
		for strings.Contains(remaining, delim) {
			idx := strings.Index(remaining, delim)
			line := remaining[:idx]
			if isFirstLine == true {
				isFirstLine = false
				NewHttpRequestHeader.InitialLine = line
			} else {
				if !strings.Contains(line, ":") {
					NewHttpRequestHeader.IsBadRequest = true
				}
				colonIdx := strings.Index(line, ":")
				key := line[:colonIdx]
				value := strings.Fields(line[colonIdx+1:])[0]
				// Check for malformed header
				for _, c := range key {
					if !unicode.IsLetter(c) {
						NewHttpRequestHeader.IsBadRequest = true
						break
					}
				}
				if key == "Host" {
					NewHttpRequestHeader.Host = value
				} else if key == "Connection" {
					NewHttpRequestHeader.Connection = value
				}
			}
			// Handle any complete requests
			remaining = remaining[idx+2:]
			if len(remaining) >= 2 && remaining[:2] == delim {
				// Update any ongoing requests
				remaining = remaining[2:]
				// Send complete request
				isFirstLine = true
				hs.handleResponse(&NewHttpRequestHeader, conn)
				NewHttpRequestHeader.InitialLine = ""
				NewHttpRequestHeader.Host = ""
				NewHttpRequestHeader.Connection = ""
			} else {
				break
			}
		}
	}
}
