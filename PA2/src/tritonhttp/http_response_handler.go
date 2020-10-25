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
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 400 Bad Request\r\n"
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleFileNotFoundRequest(requestHeader *HttpRequestHeader, conn net.Conn) {
	var NewHttpResponseHeader HttpResponseHeader
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 404 Not Found\r\n"
	hs.sendResponse(NewHttpResponseHeader, conn)
}

func (hs *HttpServer) handleResponse(requestHeader *HttpRequestHeader, conn net.Conn) (result string) {

	initialLine := requestHeader.InitialLine
	initialLineTokens := strings.Split(initialLine, " ")

	var NewHttpResponseHeader HttpResponseHeader

	if initialLineTokens[1] == "/" {

		file, err := os.Open(hs.DocRoot + "/index.html")
		if err != nil {
			hs.handleFileNotFoundRequest(requestHeader, conn)
			log.Fatal(err)
		}

		stat, err := file.Stat()
		if err != nil {
			log.Fatal(err)
		}

		// Write index.html to connection
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		// r := bufio.NewReader(file)
		// for {
		// 	bytes, err := r.ReadBytes('\n')
		// 	if err != nil {
		// 		break
		// 	}
		// 	_, err = conn.Write(bytes)
		// 	if err != nil {
		// 		return "HTTP/1.1 400 Bad Request"
		// 	}
		// }
		// file.Close()
		NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
		NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
		NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
		NewHttpResponseHeader.ContentType = hs.MIMEMap[".html"] + "\r\n"
		NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
		NewHttpResponseHeader.FilePath = hs.DocRoot + "/index.html"
		hs.sendResponse(NewHttpResponseHeader, conn)
		return "HTTP/1.1 200 OK"
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
		log.Fatal(err)
		hs.handleFileNotFoundRequest(requestHeader, conn)
	}

	stat, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}

	file.Close()
	NewHttpResponseHeader.InitialLine = "HTTP/1.1 200 OK\r\n"
	NewHttpResponseHeader.Server = "Go-Triton-Server-1.0\r\n"
	NewHttpResponseHeader.LastModified = fmt.Sprintf("%s\r\n", stat.ModTime().Format(time.RFC1123Z))
	NewHttpResponseHeader.ContentType = hs.MIMEMap[extension] + "\r\n"
	NewHttpResponseHeader.ContentLength = strconv.Itoa(int(stat.Size())) + "\r\n"
	NewHttpResponseHeader.FilePath = hs.DocRoot + initialLineTokens[1]
	hs.sendResponse(NewHttpResponseHeader, conn)
	return "HTTP/1.1 200 OK"
}

func (hs *HttpServer) sendResponse(responseHeader HttpResponseHeader, conn net.Conn) {

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

	// Send file if required
	if tokens[1] == "200" {
		buf := make([]byte, 10)
		file, err := os.Open(responseHeader.FilePath)
		if err != nil {
			log.Fatal(err)
		}
		if responseHeader.FilePath != "" {
			_, err := io.CopyBuffer(conn, file, buf)
			if err != nil {
				log.Fatal(err)
			}
		}
		file.Close()
	}

}
