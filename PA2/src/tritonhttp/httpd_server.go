package tritonhttp

import (
	"net"
)

/**
	Initialize the tritonhttp server by populating HttpServer structure
**/
func NewHttpdServer(port, docRoot, mimePath string) (*HttpServer, error) {

	// Initialize mimeMap for server to refer
	mimeMap, err := ParseMIME(mimePath)
	if err != nil {
		return nil, err
	}

	// Return pointer to HttpServer
	httpdServer := HttpServer{port, docRoot, mimePath, mimeMap}

	return &httpdServer, nil
}

/**
	Start the tritonhttp server
**/
func (hs *HttpServer) Start() (err error) {

	// Start listening to the server port
	l, err := net.Listen("tcp", hs.ServerPort)
	if err != nil {
		return err
	}
	defer l.Close()

	// Accept connection from client
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		// Spawn a go routine to handle request
		go hs.handleConnection(c)
	}

	return nil

}
