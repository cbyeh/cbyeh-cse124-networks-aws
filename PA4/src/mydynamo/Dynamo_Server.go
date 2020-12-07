package mydynamo

import (
	"errors"
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
	// If already crashed, return error
	if s.isCrashed {
		return errors.New("Error in Crash: Server is already crashed")
	}
	s.isCrashed = true
	go time.Sleep(time.Duration(seconds) * time.Second)
	s.isCrashed = false
	return nil
}

// Puts a value to the server
func (s *DynamoServer) RemotePut(value *PutArgs, result *bool) error {
	if s.isCrashed {
		*result = false
		return errors.New("Error in RemotePut: Server is crashed")
	}
	// _, ok := s.data[value.Key]
	// // If exists on DynamoServer, check if newContext < Context
	// if ok {
	// 	if value.Context.Clock.LessThan(s.data[value.Key].Context.Clock) {
	// 		*result = false
	// 		return nil
	// 	}
	// }
	// Put on this DynamoNode
	s.data[value.Key] = ObjectEntry{value.Context, value.Value}
	*result = true
	return nil
}

// Put a file to this server and W other servers
func (s *DynamoServer) Put(value PutArgs, result *bool) error {
	if s.isCrashed {
		*result = false
		return errors.New("Error in Put: Server is crashed")
	}
	// Check if newContext < oldContext, if so, keep the existing value
	_, ok := s.data[value.Key]
	if ok {
		if value.Context.Clock.LessThan(s.data[value.Key].Context.Clock) {
			*result = false
			return nil
		}
	}
	// Put on this server
	newObjectEntry := ObjectEntry{
		Context: value.Context,
		Value:   value.Value,
	}
	s.data[value.Key] = newObjectEntry
	*result = true
	// Increment vector clock element for local server
	value.Context.Clock.Increment(value.Key)
	// Replicate on the top W nodes
	for i := 0; i < s.wValue - 1 && i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]
		// Connect and see if available
		conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		var res bool
		e = conn.Call("MyDynamo.RemotePut", &value, &res)
		// If not, add to not replicated list
		if e != nil || !res {
			s.notWrittenList = append(s.notWrittenList, node)
		}
	}
	// return err
	return nil
}

// Get a file from this server
func (s *DynamoServer) RemoteGet(key string, entry *ObjectEntry) error {
	if s.isCrashed {
		return errors.New("Error in RemoteGet: Server is crashed")
	}
	// Check for causaulity, TODO:
	if _, ok := s.data[key]; ok {
		*entry = s.data[key]
	} else {
		return errors.New("Error in RemoteGet: Entry not found")
	}
	return nil
}

// Get a file from this server, matched with R other servers
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	if s.isCrashed {
		return errors.New("Error in Get: Server is crashed")
	}
	// First check local
	if _, ok := s.data[key]; ok {
		result.EntryList = append(result.EntryList, s.data[key])
	}
	for i := 0; i < s.rValue - 1 && i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]
		// Connect
		conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		if e != nil {
			return e
		}
		var res ObjectEntry
		e = conn.Call("MyDynamo.RemoteGet", key, &res)
		if e != nil {
			result.EntryList = append(result.EntryList, res)
		}
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
		data:           make(map[string]ObjectEntry),
		notWrittenList: make([]DynamoNode, 0),
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
