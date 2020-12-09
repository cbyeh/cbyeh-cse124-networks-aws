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
	data           map[string][]ObjectEntry // Key value store
	notWrittenList []DynamoNode           // All nodes that haven't been replicated
}

func (s *DynamoServer) SendPreferenceList(incomingList []DynamoNode, _ *Empty) error {
	if s.isCrashed {
		return errors.New("Error in SendPreferenceList: Server is crashed")
	}
	s.preferenceList = incomingList
	return nil
}

// Copies store from calling node to this node
func (s* DynamoServer) RemoteGossip(store map[string][]ObjectEntry, result *bool) error {
	if s.isCrashed {
		return errors.New("Error in RemoteGossip: Server is crashed")
	}
	for k, objEntryList := range store {
		for _, objEntry := range objEntryList {
			value := NewPutArgs(k, objEntry.Context, objEntry.Value)
			s.PutHelper(value, nil)
		} 
	}
	*result = true
	return nil
}

// Forces server to gossip. As this method takes no arguments, we must use the Empty placeholder
func (s *DynamoServer) Gossip(_ Empty, _ *Empty) error {
	if s.isCrashed {
		return errors.New("Error in Gossip: Server is crashed")
	}
	for i := 0; i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]
		if node == s.selfNode {
			continue
		}
		// Connect to DynamoServer
		conn, e := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
		if e != nil {
			continue
		}
		var res bool
		// Perform call
		e = conn.Call("MyDynamo.RemoteGossip", s.data, &res)
		if e != nil {
			return e
		}
	}

	return nil

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
func (s *DynamoServer) PutHelper(value PutArgs, _ *Empty) error {
	if s.isCrashed {
		return errors.New("Error in PutHelper: Server is crashed")
	}
	objList, ok := s.data[value.Key]
	hasConcurrent := false
	causallyDescended := false
	newEntryList := make([]ObjectEntry, 0)
	// If exists on DynamoServer, check if newContext < Context
	if ok {
		for i := 0; i < len(objList); i++ {
			objEntry := objList[i]
			// Check for Equals
			if value.Context.Clock.Equals(objEntry.Context.Clock) {
				hasConcurrent = true
				break
			} else if value.Context.Clock.LessThan(objEntry.Context.Clock) {
				// append to possible new entry list
				newEntryList = append(newEntryList, objEntry)
			} else if objEntry.Context.Clock.LessThan(value.Context.Clock) {
				// At least one entry is causually descended by new value, don't add
				causallyDescended = true
			}
		}
	} else {
		causallyDescended = true
	}
	if !hasConcurrent && causallyDescended {
		newEntryList = append(newEntryList, ObjectEntry{value.Context, value.Value})
		s.data[value.Key] = newEntryList
	}
	return nil
}

// Put a file to this server and W other servers
func (s *DynamoServer) Put(value PutArgs, result *bool) error {
	*result = false
	if s.isCrashed {
		return errors.New("Error in Put: Server is crashed")
	}
	// First put locally
	conn, e := rpc.DialHTTP("tcp", s.selfNode.Address+":"+s.selfNode.Port)
	if e != nil {
		return e
	}
	e = conn.Call("MyDynamo.PutHelper", value, nil)
	// Increment vector clock element for local server
	value.Context.Clock.Increment(value.Key)
	success := 1
	// Replicate on the top W nodes
	for i := 0; success < s.wValue && i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]
		// If a node sees itself on the preference list, it should be ignored
		if s.selfNode == node {
			continue
		}
		// Connect and see if available
		conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		e = conn.Call("MyDynamo.PutHelper", value, nil)
		// If not, add to not replicated list
		if e != nil {
			s.notWrittenList = append(s.notWrittenList, node)
		} else {
			success++
		}
	}
	if success == s.wValue {
		*result = true
	}
	return nil
}

// Get a file from this server
func (s *DynamoServer) GetHelper(key string, result *DynamoResult) error {
	if s.isCrashed {
		return errors.New("Error in GetHelper: Server is crashed")
	}
	objList, ok := s.data[key]
	// Add everything to result
	if ok {
		result.EntryList = append(result.EntryList, objList...)
	}
	return nil
}

// Get a file from this server, matched with R other servers
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	if s.isCrashed {
		return errors.New("Error in Get: Server is crashed")
	}
	// First check local
	success := 0
	conn, e := rpc.DialHTTP("tcp", s.selfNode.Address+":"+s.selfNode.Port)
	e = conn.Call("MyDynamo.GetHelper", key, result)
	if e == nil {
		success++
	}
	// Iterate same process for the rest of R quorums
	for i := 0; success < s.rValue && i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]
		// If a node sees itself on the preference list, it should be ignored
		if s.selfNode == node {
			continue
		}
		// Connect
		conn, e := rpc.DialHTTP("tcp", node.Address+":"+node.Port)
		e = conn.Call("MyDynamo.GetHelper", key, result)
		if e == nil {
			success++
		}
	}
	// Perform causality checks over all nodes
	objList := result.EntryList
	if len(objList) == 0 {
		return nil
	}
	latestEntry := objList[0]
	// First find most causally descended entry
	for i := 1; i < len(objList); i++ {
		objEntry := objList[i]
		if latestEntry.Context.Clock.LessThan(objEntry.Context.Clock) {
			latestEntry = objEntry
		}
	}
	// Append all concurrent to most causally descended entry
	resultList := make([]ObjectEntry, 0)
	resultList = append(resultList, latestEntry)
	for i := 0; i < len(objList); i++ {
		objEntry := objList[i]
		if latestEntry.Context.Clock.Concurrent(objEntry.Context.Clock) {
			if latestEntry.Context.Clock.Equals(objEntry.Context.Clock) {
				continue
			} else {
				resultList = append(resultList, objEntry)
			}
		} 
	}
	// Add results to parameter result
	result.EntryList = resultList
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
		data:           make(map[string][]ObjectEntry),
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
