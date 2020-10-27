package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (hs *HttpServer) handleBadRequest(conn net.Conn) {
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 400 Bad Request\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleFileNotFoundRequest(requestHeader *HttpRequestHeader, conn net.Conn) {
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 404 Not Found\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleResponse(requestHeader *HttpRequestHeader, conn net.Conn) {
	initialLineTokens := strings.Fields(requestHeader.InitialLine)

	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"

	if initialLineTokens[1][len(initialLineTokens[1])-1:] == "/" { // If last character is "/"
		file, err := os.Open(hs.DocRoot + initialLineTokens[1] + "index.html")
		if err != nil {
			hs.handleFileNotFoundRequest(requestHeader, conn)
			return
		}
		stat, _ := file.Stat()
		// Write index.html to connection
		NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
		NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
		NewHttpResponseHeader.ContentType = hs.MIMEMap[".html"] + "\r\n"
		NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
		NewHttpResponseHeader.FilePath = hs.DocRoot + initialLineTokens[1] + "index.html"
		hs.sendResponse(NewHttpResponseHeader, conn)
	}
	// Else handle non-root request
	location := hs.DocRoot + initialLineTokens[1]
	// If not a valid mime type, bad request
	extension := filepath.Ext(location)
	_, ok := hs.MIMEMap[extension]
	if !ok {
		// use MIME type application/octet-stream
	}
	file, err := os.Open(location)
	if err != nil {
		hs.handleFileNotFoundRequest(requestHeader, conn)
	}

	stat, _ := file.Stat()

	file.Close()
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
	NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
	NewHttpResponseHeader.ContentType = hs.MIMEMap[extension] + "\r\n"
	NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
	NewHttpResponseHeader.FilePath = hs.DocRoot + initialLineTokens[1]
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) sendResponse(responseHeader HttpResponseHeader, conn net.Conn) {
	// Send headers

	// Send file if required

	// Hint - Use the bufio package to write response

	// Send headers
	w := bufio.NewWriter(conn)
	response := ""
	initialLine := responseHeader.InitialLine // Append initial line first
	response += initialLine
	tokens := strings.Split(initialLine, " ")

	response += "Server: " + responseHeader.Server
	if tokens[1] == "200" { // If OK, append Last-Modified, Content-Type, Content-Length
		response += "Last-Modified: " + responseHeader.LastModified
		response += "Content-Type: " + responseHeader.ContentType
		response += "Content-Length: " + responseHeader.ContentLength
	}
	response += responseHeader.Connection
	response += "\r\n"
	println("\n" + response + "\n")
	fmt.Fprint(w, response)
	w.Flush()

	if tokens[1] == "200" {
		// Send file if required
		buf := make([]byte, 10)
		file, err := os.Open(responseHeader.FilePath)
		if err != nil {
			return
		}
		if responseHeader.FilePath != "" {
			_, err := io.CopyBuffer(conn, file, buf)
			if err != nil {
				return
			}
		}
		file.Close()
	}
}
