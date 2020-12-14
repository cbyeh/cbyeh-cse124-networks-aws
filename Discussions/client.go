package main 

import(
	"fmt"
	"log"
	"net/rpc"
)

type Item struct{
	Title string
	Body string
}

func main(){
	var reply Item
	var db []Item
	
	client, err := rpc.DialHTTP("tcp", "localhost:4040")

	if  err != nil {
		log.Fatal("Could not connect to server: ", err)
	}

	a := Item{"First", "A first item"}
	b := Item{"Second", "A second item"}
	c := Item{"Third", "A third item"}

	client.Call("API.AddItem", a, &reply) 
	// &reply means "value" stored "at" reply
	client.Call("API.AddItem", b, &reply)
	client.Call("API.AddItem", c, &reply)

	//client.Call("API.GetDB", "", &db)

	client.Call("API.GetByName", "First", &reply)

	fmt.Printf("returned item : ", reply)

	//fmt.Println("Database :", db)
}
