package mydynamo

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"errors"
	"time"
)

type DynamoServer struct {
	/*------------Dynamo-specific-------------*/
	wValue         int          //Number of nodes to write to on each Put
	rValue         int          //Number of nodes to read from on each Get
	preferenceList []DynamoNode //Ordered list of other Dynamo nodes to perform operations o
	selfNode       DynamoNode   //This node's address and port info
	nodeID         string       //ID of this node

	data		   map[string][]ObjectEntry
	isCrashed	   bool

}

func (s *DynamoServer) SendPreferenceList(incomingList []DynamoNode, _ *Empty) error {
	if s.isCrashed {
		return errors.New("Crashed in SendPreferenceList()")
	}
	s.preferenceList = incomingList
	return nil
}

func (s* DynamoServer) LocalGossip(store map[string][]ObjectEntry, result *bool) error {
	if s.isCrashed {
		// log.Println("crash in local gossip")
		return errors.New("Crashed in LocalGossip()")
	}
	// log.Println(s.selfNode.Address + ":" + s.selfNode.Port)
	for k, objEntryList := range store {

		for _, objEntry := range objEntryList {
			// Connect to DynamoServer
			conn, e := rpc.DialHTTP("tcp", s.selfNode.Address + ":" + s.selfNode.Port)
			if e != nil {
				log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in for loop Dial")
				continue
			}

			// log.Println(objEntry.Context.Clock.PairMap)
			var res bool

			value := PutArgs{k, objEntry.Context, objEntry.Value}
			
			// log.Println("calling local put from local gossip")
			// Perform call
			e = conn.Call("MyDynamo.LocalPut", value, &res)
			if e != nil {
				log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in for loop Call")
			}
			conn.Close()
		} 
	}

	*result = true
	return nil
}

// Forces server to gossip
// As this method takes no arguments, we must use the Empty placeholder
func (s *DynamoServer) Gossip(_ Empty, _ *Empty) error {
	if s.isCrashed {
		// log.Println("Gossip Crashed")
		return errors.New("Crashed in Gossip()")
	}

	for i := 0; i < len(s.preferenceList); i++ {
		node := s.preferenceList[i]

		if node == s.selfNode {
			continue
		}
		// Connect to DynamoServer
		conn, e := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
		if e != nil {
			log.Println(node.Address + ":" + node.Port + " crashed in for loop Gossip Dial")
			continue
		}

		var res bool

		// Perform call
		e = conn.Call("MyDynamo.LocalGossip", s.data, &res)
		if e != nil {
			log.Println(node.Address + ":" + node.Port + " crashed in for loop Gossip Call")
		}
		conn.Close()

	}

	return nil

}

//Makes server unavailable for some seconds
func (s *DynamoServer) Crash(seconds int, success *bool) error {
	// If already crashed, return error
	if s.isCrashed {
		*success = false
		return errors.New("Error in Crash: Server is already crashed")
	}
	// log.Println("Crashing Node " + s.selfNode.Address + ":" + s.selfNode.Port)
	go CrashTimer(s, seconds)
	*success = true
	return nil
}

func CrashTimer(s *DynamoServer, seconds int) {
	s.isCrashed = true
	time.Sleep(time.Duration(seconds) * time.Second)
	s.isCrashed = false
	// log.Println("Crash Done " + s.selfNode.Address + ":" + s.selfNode.Port)
	return
}

// Puts a value to the server.
// Check causality
func (s *DynamoServer) LocalPut(value PutArgs, result *bool) error {
	if s.isCrashed {
		*result = false
		// log.Println("Local Put Crashed")
		return errors.New("Crashed in LocalPut()")
	}

	objEntryList, ok := s.data[value.Key]

	newEntryList := make([]ObjectEntry, 0)

	isLessThan := false
	containsNew := false
	casuallyDescended := false

	concurrentCount := 0

	// log.Println()
	// log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " Current EntryList")
	// log.Println(objEntryList)
	
	// log.Println("\nNew Put args clock")
	// log.Println(value.Context.Clock.PairMap)

	// If exists on DynamoServer, local check for casuality
	if ok {
		for i := 0; i < len(objEntryList); i++ {
			objEntry := objEntryList[i]
			// log.Println(objEntry.Context.Clock.PairMap)
			// Check for Equals
			if value.Context.Clock.Equals(objEntry.Context.Clock) {
				// New context already in node, keep existing values
				// log.Println("Same context")
				// log.Println(string(objEntry.Value))
				containsNew = true
				break
			} else if value.Context.Clock.LessThan(objEntry.Context.Clock) {
				// Less than -> automatically keep all existing values
				// log.Println("Less than")
				// log.Println(string(objEntry.Value))
				isLessThan = true
				break
			} else if objEntry.Context.Clock.LessThan(value.Context.Clock) {
				// At least one entry is causually descended by new value
				// log.Println("Greater than")
				// log.Println(string(objEntry.Value))
				casuallyDescended = true
			} else if value.Context.Clock.Concurrent(objEntry.Context.Clock) {
				// append to possible new entry list
				// log.Println("concurrent ")
				// log.Println(value.Context)
				// log.Println(objEntry.Context)
				// log.Println(string(objEntry.Value))
				newEntryList = append(newEntryList, objEntry)
				concurrentCount += 1
			}
		}
	} else {
		casuallyDescended = true
	}

	if !containsNew && !isLessThan && (casuallyDescended || concurrentCount == len(objEntryList)) {
		newEntryList = append(newEntryList, ObjectEntry{value.Context, value.Value})
		
		s.data[value.Key] = newEntryList
		// log.Println("APPEND NEW CONTEXT")
	} 
	// log.Println()
	// log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " EntryList")
	
	*result = true
	return nil
}

// Puts a new PutArgs to on remote server
func (s *DynamoServer) Put(value PutArgs, result *bool) error {
	if s.isCrashed {
		*result = false
		// log.Println("Put Crashed")
		return errors.New("Crashed in Put()")
	}
	// log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " called Put")
	// Increment key in vector clock
	value.Context.Clock.Increment(s.nodeID)


	// Connect to DynamoServer
	conn, e := rpc.DialHTTP("tcp", s.selfNode.Address + ":" + s.selfNode.Port)
	if e != nil {
		log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in Dial")
		// s.notVisited = append(s.notVisited, node)
		// continue
	}

	// log.Println(value.Context.Clock.PairMap)
	var res bool

	// Perform call
	e = conn.Call("MyDynamo.LocalPut", value, &res)
	if e != nil {
		log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in  Call")
		// s.notVisited = append(s.notVisited, node)
	}
	conn.Close()
	success := 1
	
	for i := 0; i < len(s.preferenceList); i++ {
		if success == s.wValue {
			// log.Println("done")
			// If done, append remaining preference list nodes into list
			// for Gossip
			// s.notVisited = append(s.notVisited, s.preferenceList[i:]...)
			break
		}
		node := s.preferenceList[i]
		if node == s.selfNode {
			continue
		}
		// log.Println("Replicate Put on " + node.Address + ":" + node.Port)
		
		// Connect to DynamoServer
		conn, e := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
		if e != nil {
			log.Println(node.Address + ":" + node.Port + " crashed in for loop Dial")
			continue
		}

		// log.Println(value.Context.Clock.PairMap)
		var res bool

		// Perform call
		e = conn.Call("MyDynamo.LocalPut", value, &res)
		if e != nil {
			log.Println(node.Address + ":" + node.Port + " crashed in for loop Call")
		} else {
			success += 1
		}
		conn.Close()
	}

	if success == s.wValue {
		*result = true
	} else {
		*result = false
	}

	return nil

}

func (s *DynamoServer) LocalGet(key string, result *[]ObjectEntry) error {
	if s.isCrashed {
		*result = nil
		return errors.New("Crashed in LocalGet()")
	}

	objEntryList, ok := s.data[key]
	if !ok {
		*result = nil
		return nil
	}
	*result = objEntryList
	return nil

}

//Get a file from this server, matched with R other servers
// Checks for causalities
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	var res DynamoResult
	newEntryList := make([]ObjectEntry, 0)
	
	if s.isCrashed {
		res.EntryList = newEntryList
		*result = res
		return errors.New("Crashed in Get()")
	}
	// log.Println("GET")
	// log.Println("Get local " + s.selfNode.Address + ":" + s.selfNode.Port)
	conn, e := rpc.DialHTTP("tcp", s.selfNode.Address + ":" + s.selfNode.Port)
	if e != nil {
		log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in Dial")
		
	}

	var objEntryList []ObjectEntry
	e = conn.Call("MyDynamo.LocalGet", key, &objEntryList)
	if e != nil {
		// *result = nil
		log.Println(s.selfNode.Address + ":" + s.selfNode.Port + " crashed in Call")
	}

	newEntryList = append(newEntryList, GetHelper(objEntryList)...)
	success := 1

	for i := 0; i < len(s.preferenceList); i++ {
		if success == s.rValue {
			break
		}
		node := s.preferenceList[i]
		if node == s.selfNode {
			continue
		}

		// log.Println(node.Address + ":" + node.Port)
		conn, e := rpc.DialHTTP("tcp", node.Address + ":" + node.Port)
		if e != nil {
			continue
		}

		var currObjEntryList []ObjectEntry
		e = conn.Call("MyDynamo.LocalGet", key, &currObjEntryList)
		if e != nil {
			// *result = nil
			continue
		}
		
		GetHelper(currObjEntryList)

		for _, objEntryToAppend := range currObjEntryList {
			if !ContainsObjectEntry(newEntryList, objEntryToAppend) {
				newEntryList = append(newEntryList, objEntryToAppend)
			}
		}
		success += 1
	}

	newEntryList = GetHelper(newEntryList)
	res.EntryList = newEntryList
	*result = res

	// printObjectEntryList(newEntryList)
	return nil
}

func ContainsObjectEntry(entryList []ObjectEntry, objEntry ObjectEntry) bool {
	for _, currObjEntry := range entryList {
		if currObjEntry.Context.Clock.Equals(objEntry.Context.Clock) {
			return true
		}
	}
	return false

}

func GetHelper(objEntryList []ObjectEntry) []ObjectEntry {
	newEntryList := make([]ObjectEntry, 0)

	if objEntryList != nil && len(objEntryList) > 0 {
		bestObjEntry := objEntryList[0]
		// bestInd := 0
		for i := 1; i < len(objEntryList); i++ {
			objEntry := objEntryList[i]
			if bestObjEntry.Context.Clock.LessThan(objEntry.Context.Clock) {
				bestObjEntry = objEntry
				// bestInd = i
			}

		}
		// objEntryList = remove(objEntryList, bestInd)
		newEntryList = append(newEntryList, bestObjEntry)

		for i := 0; i < len(objEntryList); i++ {
			objEntry := objEntryList[i]
			if objEntry.Context.Clock.Equals(bestObjEntry.Context.Clock) {
				continue
			}
			if bestObjEntry.Context.Clock.Concurrent(objEntry.Context.Clock) {
				newEntryList = append(newEntryList, objEntry)
			}
		}
	}

	return newEntryList

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
		data:			make(map[string][]ObjectEntry),
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
