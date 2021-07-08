package main

import (
	"fmt"
	"time"
	"math/rand"
	"strconv"
	"strings"
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
	// fmt.Printf("server has slept for %v seconds\n", amt)
}
func SleepT(){
	amt := time.Duration(1000)
    time.Sleep(time.Millisecond * amt)
}
func Max(x, y int) int {
    if x < y {
        return y
    }
    return x
}
func returnNewTime(clientClock int, serverClock int) int{
	return Max(clientClock,serverClock)+1
}
func logTime(server_time int, msgTime int, clientId int,server bool) (int, string) {
	//bool is the flag where server -> client for True
	server_time = returnNewTime(server_time, msgTime)
	server_time_s := strconv.Itoa(server_time)
	msg_s := strconv.Itoa(msgTime)
	client_s := strconv.Itoa(clientId)
	var log_time string
	if server == false{
		log_time = client_s+ "."+ msg_s + ",S." + server_time_s
	}else{
		log_time = "S." + server_time_s + "," + client_s+ ".?"
	}
	return server_time, log_time
}
func OrderHelper(time int, clientId int, server bool) (string) {
	//bool is the flag where server -> client for True
	time_s := strconv.Itoa(time)
	client_s := strconv.Itoa(clientId)
	var log_time string
	if server == false{
		log_time = client_s+ "." + time_s
	}else{
		log_time = "S." + time_s
	}
	return log_time
}
func totalOrder(logs[]string, no_client int) []string {
	var totalOrderLogs []string
	var no_client_clock_track []int = make([]int,no_client)
	var server_clock_track int
	const s_constant string ="s"
	fmt.Printf("logshere %v",logs)
	// fmt.Println("helloooooo")
	for i:=0;i<no_client;i++{
		no_client_clock_track[i]=0
	}
	for log := range logs {
		// fmt.Printf("logfadfadf %v",log)
		indivClock := strings.Split(logs[log],",")
		// fmt.Printf("indivClock %v",indivClock)
		// fmt.Printf("indivClock[0] %v",indivClock[0])
		sender := strings.Split(indivClock[0],".")
		sender_id := sender[0]
		sender_time :=sender[1]
		// fmt.Printf("sender %v",sender[1])
		// receiver := strings.Split(indivClock[1],".")
		if sender_id!=s_constant {
			// fmt.Printf("sender[0] %v\n",sender_id)
			sender_id,_ := strconv.Atoi(sender_id)
			// print("b")
			sender_clock,_ := strconv.Atoi(sender_time)
			// print("a")
			receiver_clock,_ := strconv.Atoi(sender[1])
			last_seen_client_clock:=no_client_clock_track[sender_id]
			// print("c")
			for i:=last_seen_client_clock;i<sender_clock+1;i++{
				newEntry:= OrderHelper(i,sender_id,false)
				totalOrderLogs= append(totalOrderLogs,newEntry)
			}
			last_seen_server_clock:=server_clock_track
			for i:=last_seen_server_clock;i<receiver_clock+1;i++{
				newEntry:= OrderHelper(i,0,true)
				totalOrderLogs= append(totalOrderLogs,newEntry)
			}
		}else{
			sender_clock,_ := strconv.Atoi(sender[1])
			last_seen_server_clock:=server_clock_track
			for i:=last_seen_server_clock;i<sender_clock+1;i++{
				newEntry:= OrderHelper(i,0,true)
				totalOrderLogs= append(totalOrderLogs,newEntry)
			}

		}
		
		
	}	
	return totalOrderLogs
}


//////////////////////////////////
/*client-server functions*////////
//////////////////////////////////
func client(id int, clock_client_to_server chan int, clock_server_to_client chan int) {
	/*
	@id int an identifying number for client node
	@clock chan int sends an int over the chan
	*/
	var client_clock int = 1
	fmt.Printf("Client %v sent time %v to server\n", id, client_clock)
    clock_client_to_server <- client_clock
	for {
		// receive message from clock1 else check how many broadcast messages has client received
		select {
      		case msg1 := <- clock_server_to_client:
				fmt.Printf("Client %v receives %v from Server. Client clock compares msg %v and localClock %v.\n", id, msg1,msg1,client_clock)
				client_clock = returnNewTime(msg1,client_clock)
        		fmt.Printf("Client clock is now %v.%v\n",id, client_clock)
				SleepT()
			case <- time.After(time.Second):
				client_clock = client_clock+1
				fmt.Printf("Client clock is now %v.%v\n",id,client_clock)
				SleepT()
				clock_client_to_server <- client_clock
    	}
  	}
}


func server(clock_client_to_server [no_client]chan int, clock_server_to_client [no_client]chan int){
	var server_time int = 2
	var logs []string
	var log_client_msg_s string
	var log_server_msg_s string
	for {
		// receive message from clock1 or clock2 whichever is ready
		// if both clock1 and clock2 are ready, then break the tie randomly	
		select {
      		case msg1 := <- clock_client_to_server[0]:
        		fmt.Printf("Server receiving %v from client 0\n", msg1)
				server_time, log_client_msg_s = logTime(server_time,msg1,0,false)
				logs = append(logs,log_client_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				go serverSendOneClientTime(clock_server_to_client[0],0,server_time)
				server_time, log_server_msg_s = logTime(server_time,0,0,true)
				logs = append(logs,log_server_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				goSleep()
      		case msg2 := <- clock_client_to_server[1]:
        		fmt.Printf("Server receiving %v from client 1\n", msg2)
				server_time, log_client_msg_s = logTime(server_time,msg2,1,false)
				logs = append(logs,log_client_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				go serverSendOneClientTime(clock_server_to_client[1],1,server_time)
				server_time, log_server_msg_s = logTime(server_time,0,1,true)
				logs = append(logs,log_server_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				goSleep()
			case msg3 := <- clock_client_to_server[2]:
        		fmt.Printf("Server receiving %v from client 2\n", msg3)
				server_time, log_client_msg_s = logTime(server_time,msg3,2,false)
				logs = append(logs,log_client_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				go serverSendOneClientTime(clock_server_to_client[2],2,server_time)
				server_time, log_server_msg_s = logTime(server_time,0,2,true)
				logs = append(logs,log_server_msg_s)
				fmt.Printf("Server clock is now S. %v\n",server_time)
				goSleep()
			case <- time.After(time.Second):
				total_order_logs := totalOrder(logs,len(clock_client_to_server))
				fmt.Printf("current logs: %v\n",logs)
				fmt.Printf("total order logs: %v\n",total_order_logs)
				server_time = server_time + 1
				fmt.Printf("Server clock is now S.%v\n",server_time)
				fmt.Println("timeout\n")
				goSleep()
    	}
  	}
}


func serverSendOneClientTime(clock_server_to_client chan int, send_client int, newClock int) int{
	clock_server_to_client <- newClock
	fmt.Printf("server msg sends updated time %v send_client %v\n", newClock, send_client)
	return newClock
}



// func serverBroadcast(clock_server_to_client [no_client]chan int, send_client int, msg int){
// 	fmt.Printf("serVer msg is running for send_client %v\n", send_client)
// 	for i := 0; i < len(clock_server_to_client); i++{
// 		if i!= send_client{
// 			clock_server_to_client[i] <- msg
// 			fmt.Printf("Server is Sending %v on behalf of client %v to client %v\n", msg,send_client,i)
// 		}
// 	}
// }

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
func testOnePointTwo(){
	var clock_client_to_server [no_client]chan int 
	var clock_server_to_client [no_client]chan int
	for i:=0; i<no_client; i++  {
		clock_client_to_server[i] = make(chan int)
		clock_server_to_client[i] = make(chan int)
	}

	go client(0, clock_client_to_server[0], clock_server_to_client[0])
	go client(1, clock_client_to_server[1], clock_server_to_client[1])
	go client(2, clock_client_to_server[2], clock_server_to_client[2])
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
	// testOnePointOne()
	testOnePointTwo()
	// erm :=strings.Split("1.2.3",".")
	// fmt.Printf("erm %v",erm)


	var input string
	fmt.Scanln(&input)
}
