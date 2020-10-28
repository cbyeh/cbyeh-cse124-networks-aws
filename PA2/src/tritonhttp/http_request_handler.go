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

	const timeout = 5 * time.Second
	const delim = "\r\n"
	isFirstLine := true
	remaining := ""

	var NewHttpRequestHeader HttpRequestHeader
	NewHttpRequestHeader.InitialLine = ""
	NewHttpRequestHeader.Host = ""
	NewHttpRequestHeader.Connection = ""
	NewHttpRequestHeader.IsBadRequest = false
	NewHttpRequestHeader.IsPartialRequest = false

	defer log.Println("Closed connection")

	// Start a loop for reading requests continuously
	for {

		// Set a timeout for read operation and read from connection
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err, ok := err.(net.Error); ok && err.Timeout() {
			NewHttpRequestHeader.IsPartialRequest = true
			hs.handleResponse(&NewHttpRequestHeader, conn)
			break
		}
		if size > 0 {
			conn.SetReadDeadline(time.Now().Add(timeout))
		} else {
			conn.Close()
			return
		}

		// Read from the connection socket into a buffer
		data := buf[:size]
		remaining += string(data)

		// Validate the request lines that were read
		for strings.Contains(remaining, delim) {
			idx := strings.Index(remaining, delim)
			line := remaining[:idx]
			key := ""
			value := ""
			if isFirstLine == true {
				isFirstLine = false
				NewHttpRequestHeader.InitialLine = line
			} else {
				if strings.Contains(line, ":") {
					colonIdx := strings.Index(line, ":")
					key = line[:colonIdx]
					var valueIdx int
					for i := colonIdx + 1; i < len(line); i++ {
						if line[i] != ' ' {
							valueIdx = i
							break
						}
					}
					value = line[valueIdx:]
				} else {
					NewHttpRequestHeader.IsBadRequest = true
				}
				// Check for malformed header
				// TODO:
				if key == "Host" {
					NewHttpRequestHeader.Host = value
				} else if key == "Connection" {
					NewHttpRequestHeader.Connection = value
				}
			}
			// Handle any complete requests
			remaining = remaining[idx+2:]
			// Finished reading request
			log.Println("Finished reading request")
			if len(remaining) >= 2 && remaining[:2] == delim {
				// Update any ongoing requests
				remaining = remaining[2:]
				// Send complete request
				hs.handleResponse(&NewHttpRequestHeader, conn)
				if NewHttpRequestHeader.Connection == "close" {
					conn.Close()
					return
				}
				// msg := <-t:
				isFirstLine = true
				NewHttpRequestHeader.InitialLine = ""
				NewHttpRequestHeader.Host = ""
				NewHttpRequestHeader.Connection = ""
				NewHttpRequestHeader.IsBadRequest = false
			}
		}
	}
}
