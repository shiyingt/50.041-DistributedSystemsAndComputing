package main

import (
	"fmt"
	"time"
	"math/rand"
)
const (
	no_client int = 5
)

/////////////////////////////////
/*helper functions*//////////////
/////////////////////////////////
func goSleep(){
	amt := time.Duration(rand.Intn(1000))
    time.Sleep(time.Millisecond * amt)
	fmt.Printf("server has slept for %v seconds\n", amt)
}
func SleepT(){
	amt := time.Duration(1000)
    time.Sleep(time.Millisecond * amt)
}

//////////////////////////////////
/*client-server functions*////////
//////////////////////////////////
func client(id int, clock_client_to_server chan int, clock_server_to_client chan int) {
	/*
	@id int an identifying number for client node
	@clock chan int sends an int over the chan
	*/
	var receiveList [no_client]int
	send := id
    fmt.Printf("Client %v sent %v to server\n", id, send)
    clock_client_to_server <- send
	amt := time.Duration(1000)
    time.Sleep(time.Millisecond * amt)
	fmt.Printf("client %v awakes\n ",id)
	receiveList[id]=id
	for {
		// receive message from clock1 else check how many broadcast messages has client received
		select {
      		case msg1 := <- clock_server_to_client:
        		fmt.Printf("Client receives %v from Server 0\n", msg1)
				receiveList[msg1]=msg1
				SleepT()
			case <- time.After(time.Second):
				fmt.Printf("Client %v has received %v\n",id,receiveList)
				SleepT()
    	}
  	}
}


func server(clock_client_to_server [no_client]chan int, clock_server_to_client [no_client]chan int){
	for {
		// receive message from clock1 or clock2 whichever is ready
		// if both clock1 and clock2 are ready, then break the tie randomly	
		select {
      		case msg1 := <- clock_client_to_server[0]:
        		fmt.Printf("Server receiving %v from client 0\n", msg1)
				go serverMessage(clock_server_to_client, 0, msg1)
				goSleep()
      		case msg2 := <- clock_client_to_server[1]:
        		fmt.Printf("Server receiving %v from client 1\n", msg2)
				go serverMessage(clock_server_to_client, 1, msg2)
				goSleep()
			case msg3 := <- clock_client_to_server[2]:
        		fmt.Printf("Server receiving %v from client 2\n", msg3)
				go serverMessage(clock_server_to_client, 2, msg3)
				goSleep()
			case msg4 := <- clock_client_to_server[3]:
        		fmt.Printf("Server receiving %v from client 3\n", msg4)
				go serverMessage(clock_server_to_client, 3, msg4)
				goSleep()
			case msg5 := <- clock_client_to_server[4]:
        		fmt.Printf("Server receiving %v from client 4\n", msg5)
				go serverMessage(clock_server_to_client, 4, msg5)
				goSleep()
			case <- time.After(time.Second):
				fmt.Println("timeout\n")
				goSleep()
    	}
  	}
}

func serverMessage(clock_server_to_client [no_client]chan int, send_client int, msg int){
	fmt.Printf("serVer msg is running for send_client %v\n", send_client)
	for i := 0; i < len(clock_server_to_client); i++{
		if i!= send_client{
			clock_server_to_client[i] <- msg
			fmt.Printf("Server is Sending %v on behalf of client %v to client %v\n", msg,send_client,i)
		}
	}
}

func testOnePointOne(){
	var clock_client_to_server [no_client]chan int 
	var clock_server_to_client [no_client]chan int
	for i:=0; i<no_client; i++  {
		clock_client_to_server[i] = make(chan int)
		clock_server_to_client[i] = make(chan int)
	}

	go client(0, clock_client_to_server[0], clock_server_to_client[0])
	go client(1, clock_client_to_server[1], clock_server_to_client[1])
	go client(2, clock_client_to_server[2], clock_server_to_client[2])
	go client(3, clock_client_to_server[3], clock_server_to_client[3])
	go client(4, clock_client_to_server[4], clock_server_to_client[4])
	go server(clock_client_to_server,clock_server_to_client)
}

func main() {
	// create a channel of type integer
	// var clock_client_to_server [no_client]chan int 
	// var clock_server_to_client [no_client]chan int
	// for i:=0; i<no_client; i++  {
	// 	clock_client_to_server[i] = make(chan int)
	// 	clock_server_to_client[i] = make(chan int)
	// }

	// go client(0, clock_client_to_server[0], clock_server_to_client[0])
	// go client(1, clock_client_to_server[1], clock_server_to_client[1])
	// go client(2, clock_client_to_server[2], clock_server_to_client[2])
	// go client(3, clock_client_to_server[3], clock_server_to_client[3])
	// go client(4, clock_client_to_server[4], clock_server_to_client[4])
	// go server(clock_client_to_server,clock_server_to_client)
	testOnePointOne()


	var input string
	fmt.Scanln(&input)
}
