package tritonhttp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (hs *HttpServer) handleBadRequest(conn net.Conn) {
	log.Println("Handled bad request")
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 400 Bad Request\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	defer conn.Close()
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleFileNotFoundRequest(conn net.Conn) {
	log.Println("Handled file not found request")
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 404 Not Found\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleResponse(requestHeader *HttpRequestHeader, conn net.Conn) {

	// Check if headers are OK
	if requestHeader.Host == "" || requestHeader.IsBadRequest == true || requestHeader.IsPartialRequest == true {
		hs.handleBadRequest(conn)
		return
	}

	var NewHttpResponseHeader HttpResponseHeader

	// Check if initial line is valid. Send good response if so and file is valid
	line := requestHeader.InitialLine
	initialLineTokens := strings.Fields(line)

	tokens := strings.Fields(line)

	// Check if initial line is valid
	if len(tokens) != 3 {
		hs.handleBadRequest(conn)
		return
	}
	if line[:4] != "GET " || tokens[1][:1] != "/" || tokens[2] != "HTTP/1.1" {
		hs.handleBadRequest(conn)
		return
	}

	// Else valid initial line and requesting root page
	log.Println("Valid initial line")
	if initialLineTokens[1][len(initialLineTokens[1])-1:] == "/" { // If last character is "/"
		file, err := os.Open(hs.DocRoot + initialLineTokens[1] + "index.html")
		if err != nil {
			hs.handleFileNotFoundRequest(conn)
			return
		}
		stat, _ := file.Stat()
		// Write index.html to connection
		NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
		NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
		NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
		NewHttpResponseHeader.ContentType = hs.MIMEMap[".html"] + "\r\n"
		NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
		NewHttpResponseHeader.FilePath = hs.DocRoot + initialLineTokens[1] + "index.html"
		hs.sendResponse(NewHttpResponseHeader, conn)
		return
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
		hs.handleFileNotFoundRequest(conn)
		return
	}
	stat, _ := file.Stat()
	file.Close()
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
	NewHttpResponseHeader.ContentType = hs.MIMEMap[extension] + "\r\n"
	NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
	NewHttpResponseHeader.FilePath = hs.DocRoot + initialLineTokens[1]
	hs.sendResponse(NewHttpResponseHeader, conn)

}

func (hs *HttpServer) sendResponse(responseHeader HttpResponseHeader, conn net.Conn) {
	log.Println("Sending response")
	// Send headers
	w := bufio.NewWriter(conn)
	response := ""
	initialLine := responseHeader.InitialLine // Append initial line first
	response += initialLine
	response += "Server: " + responseHeader.Server

	tokens := strings.Split(initialLine, " ")

	if tokens[1] == "200" { // If OK, append Last-Modified, Content-Type, Content-Length
		response += "Last-Modified: " + responseHeader.LastModified
		response += "Content-Type: " + responseHeader.ContentType
		response += "Content-Length: " + responseHeader.ContentLength
	}
	response += responseHeader.Connection
	response += "\r\n"
	fmt.Fprint(w, response)
	w.Flush()

	if tokens[1] == "200" {
		// Send file if required
		buf := make([]byte, 1024)
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
