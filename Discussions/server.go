package main

import (
	"fmt"
	"log"
	"net/rpc"
)

type Item struct{
	title string
	body string
}

type API int

var dabase []Item

func (a *API) AddItem(item Item, reply *Item) error{
	database = append(database, item)
	*reply = item
	return nil
}

func (a *API) GetByName(title string, reply *Item) error{
	var getItem Item
	
	for _, val := range database{
		if val.title == title {
			getItem = val
		}
	}

	*reply = getItem // put the Item into "reply" address; 
			//RPC stub will grab from that address and send param.
			//back to client

	return nil
}

func (a *API) GetDB(title string, reply *[]Item) error{
	*reply = database
	return nil
}

func main(){
	var api = new(API)
	err := rpc.Register(api)

	if  err != nil {
		log.Fatal("error registering API", error)
	}

	rpc.HandleHTTP()

	listener, err := net.Listen("tcp",":4040")

	if err != nil{
		log.Fatal("Could not listen on TCP port 4040", err)
	}

	log.Printf("serving rpc on port %d", 4040)
	err = http.Serve(listener, nil)
	if err != nil{
		log.Fatal("could not start listening on http: ", err)
	}
}





































