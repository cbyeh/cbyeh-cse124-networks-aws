package mydynamo

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type DynamoServer struct {
	wValue         int                    // Number of nodes to write to on each Put
	rValue         int                    // Number of nodes to read from on each Get
	preferenceList []DynamoNode           // Ordered list of other Dynamo nodes to perform operations on
	selfNode       DynamoNode             // This node's address and port info
	nodeID         string                 // ID of this node
	isCrashed      bool                   // If Crash() was called, emulate as if a node crashed
	data           map[string]ObjectEntry // Key value store
	notWrittenList []DynamoNode           // All nodes that haven't been replicated
}

func (s *DynamoServer) SendPreferenceList(incomingList []DynamoNode, _ *Empty) error {
	s.preferenceList = incomingList
	return nil
}

// Forces server to gossip
// As this method takes no arguments, we must use the Empty placeholder
func (s *DynamoServer) Gossip(_ Empty, _ *Empty) error {
	panic("todo")
}

// Makes server unavailable for some seconds
func (s *DynamoServer) Crash(seconds int, success *bool) error {
	s.isCrashed = true
	go time.Sleep(time.Duration(seconds) * time.Second)
	s.isCrashed = false
	return nil
}

// Put a file to this server and W other servers
func (s *DynamoServer) Put(value PutArgs, result *bool) error {
	// Check if newContext < oldContext, if so, keep the existing value
	if value.Context.Clock.LessThan(s.data[value.Key].Context.Clock) {
		*result = false
		return nil
	}
	// Put on this server
	newObjectEntry := ObjectEntry{
		Context: value.Context,
		Value:   value.Value,
	}
	s.data[value.Key] = newObjectEntry
	// Replicate on the top W nodes, TODO: Maybe try just for W = 1
	for i := 0; i < s.wValue; i++ {
		// node := s.preferenceList[i]
		// Connect and see if available
		// conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		// if e != nil {
		// 	*result = false
		// 	return e
		// }
		// e = conn.Call("Server.GetBlock", blockHash, block)
		// // TODO:
		// If not, add to not replicated list
	}
	// Increment vector clock element
	// value.Context.Clock.pairArray[value.Key]++
	// Store all nodes that weren't replicated
	*result = true
	return nil
}

// Get a file from this server, matched with R other servers
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	// First check local
	if val, ok := s.data[key]; ok {
		newObjectEntry := ObjectEntry{
			Context: val.Context,
			Value:   val.Value,
		}
		result.EntryList = append(result.EntryList, newObjectEntry)
	}
	// TODO: Maybe try just for R = 1
	for i := 0; i < s.rValue; i++ {
		node := s.preferenceList[i]
		// Connect
		conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		if e != nil {
			return e
		}
		println(conn)
	}
	return nil
}

/* Belows are functions that implement server boot up and initialization */
func NewDynamoServer(w int, r int, hostAddr string, hostPort string, id string) DynamoServer {
	preferenceList := make([]DynamoNode, 0)
	selfNodeInfo := DynamoNode{
		Address: hostAddr,
		Port:    hostPort,
	}
	return DynamoServer{
		wValue:         w,
		rValue:         r,
		preferenceList: preferenceList,
		selfNode:       selfNodeInfo,
		nodeID:         id,
		isCrashed:      false,
	}
}

func ServeDynamoServer(dynamoServer DynamoServer) error {
	rpcServer := rpc.NewServer()
	e := rpcServer.RegisterName("MyDynamo", &dynamoServer)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Name Registration")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Registered the RPC Interfaces")

	l, e := net.Listen("tcp", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Port Listening")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Listening to Target Port ", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	log.Println(DYNAMO_SERVER, "Serving Server Now")

	return http.Serve(l, rpcServer)
}
