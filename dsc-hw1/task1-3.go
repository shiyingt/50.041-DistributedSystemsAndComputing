package main

import (
	"fmt"
	"time"
	"math/rand"
)
const (
	no_client int = 3
)



/////////////////////////////////
/*helper functions*//////////////
/////////////////////////////////
func goSleep(){
	amt := time.Duration(rand.Intn(1000))
    time.Sleep(time.Millisecond * amt)
	// fmt.Printf("server has slept for %v seconds\n", amt)
}
func TimeLagOnOneServer(){
	amt := time.Duration(rand.Intn(300))
    time.Sleep(time.Millisecond * amt)
	// fmt.Printf("server has slept for %v seconds\n", amt)
}
func SleepT(){
	amt := time.Duration(30000)
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
func updateVectorClock(incoming_message [no_client]int, client_clock[no_client]int, clients_own_id int)[no_client]int{
	var output [no_client]int 
	for i:=range incoming_message{
		output[i]=Max(incoming_message[i],client_clock[i])
	}
	output[clients_own_id]= output[clients_own_id] +1
	return output
}
func detectCausalityViolation(incoming_message [no_client]int,client_clock[no_client]int,incoming_id int)bool{
	if incoming_message[incoming_id]<client_clock[incoming_id]{
		return true
	}
	return false
}


//////////////////////////////////
/*client-server functions*////////
//////////////////////////////////
// In the vector clock diagrams, clients send vector clocks to each other.
// The way i package client's message is vector_clock + id_of_owner_who_sent_message.This helps me identify the owner of the 
// vector clock message, such that the length of the message is length of vector clock +1. This is under the assumption that 
// id_of_owner_who_sent_message are in single digits for the current implementation.
// this implementation fails becuase of the way i implemented the sending of channels + blocking call which essentially 
// queues up the messages and the messages come sequentially based on the order that they were sent. :(
// have to go with the more complex one that is faithful to the original.
// Another way i can identify who sent the vector clock would be to have [] []chan []int where each client-to-client 
// interaction will have their own private channel. I did not chose this implementation due to the complexity of this solution.
// gg it seems i have to array [] []chan []int so i can delay particulat channels
 
func client(id int, clock_for_all  [no_client] [no_client]chan [no_client]int) {
	/*
	@id int an identifying number for client node
	@clock_for_all [] []chan []int contains a list of channels which would be used to send vector clocks to other clients.
	*/
	var vector_clock [no_client] int 
	for i:= range vector_clock {
		vector_clock[i]=0
	}
	//update clock timing of local clock
	vector_clock[id]=1
	fmt.Printf("Client %v clock %v\n", id, vector_clock)
	for {
		// receive message from clock1 else check how many broadcast messages has client received
		// you can understand clock_for_all[0] as the pigeonhole for Client 1. we have to check the different 1-1 channels
		// in the pigeonhole to see if any clients have sent stuff to client 1
		select {
			case msg1 := <- clock_for_all[id][0]:
				// if id ==0{
				// 	TimeLagOnOneServer()
				// }
				fmt.Printf("Client %v receives %v from Client 0. Client clock compares incoming vector clock %v and localClock %v.\n", id, msg1,  msg1,vector_clock)
				causality_violation:= detectCausalityViolation(msg1,vector_clock,0)
				if causality_violation{
					fmt.Printf("CAUSALITY VIOLATION: client %v received vector_clock %v from Client 0 when local clock is %v\n", id,msg1,vector_clock)
				}
				vector_clock = updateVectorClock(msg1,vector_clock,id)
        		// fmt.Printf("Client %v clock is now %v\n",id, vector_clock)
				// goSleep()
			case msg2 := <- clock_for_all[id][1]:
				// if id ==0{
				// 	TimeLagOnOneServer()
				// }
				fmt.Printf("Client %v receives %v from Client 1. Client clock compares incoming vector clock %v and localClock %v.\n", id, msg2,  msg2,vector_clock)
				causality_violation:= detectCausalityViolation(msg2,vector_clock,1)
				if causality_violation{
					fmt.Printf("CAUSALITY VIOLATION: client %v received vector_clock %v from Client 1 when local clock is %v\n", id,msg2,vector_clock)
				}
				vector_clock = updateVectorClock(msg2,vector_clock,id)
        		// fmt.Printf("Client %v clock is now %v\n",id, vector_clock)
				// goSleep()
			case msg3 := <- clock_for_all[id][2]:
				// if id ==0{
				// 	TimeLagOnOneServer()
				// }
				fmt.Printf("Client %v receives %v from Client2. Client clock compares incoming vector clock %v and localClock %v.\n", id, msg3,  msg3,vector_clock)
				causality_violation:= detectCausalityViolation(msg3,vector_clock,2)
				if causality_violation{
					fmt.Printf("CAUSALITY VIOLATION: client %v received vector_clock %v from Client 2 when local clock is %v\n", id,msg3,vector_clock)
				}
				vector_clock = updateVectorClock(msg3,vector_clock,id)
        		// fmt.Printf("Client %v clock is now %v\n",id, vector_clock)
				goSleep()
			// case msg4 := <- clock_for_all[id][3]:
			// 	// if id ==0{
			// 	// 	TimeLagOnOneServer()
			// 	// }
			// 	fmt.Printf("Client %v receives %v from Client3. Client clock compares incoming vector clock %v and localClock %v.\n", id, msg4,  msg4,vector_clock)
			// 	causality_violation:= detectCausalityViolation(msg4,vector_clock,3)
			// 	if causality_violation{
			// 		fmt.Printf("CAUSALITY VIOLATION: client %v received vector_clock %v from Client 3 when local clock is %v\n", id,msg4,vector_clock)
			// 	}
			// 	vector_clock = updateVectorClock(msg4,vector_clock,id)
        	// 	fmt.Printf("Client %v clock is now %v\n",id, vector_clock)
			// 	// goSleep()
			// case msg5 := <- clock_for_all[id][4]:
			// 	// if id ==0{
			// 	// 	TimeLagOnOneServer()
			// 	// }
			// 	fmt.Printf("Client %v receives %v from Client4. Client clock compares incoming vector clock %v and localClock %v.\n", id, msg5,  msg5,vector_clock)
			// 	causality_violation:= detectCausalityViolation(msg5,vector_clock,4)
			// 	if causality_violation{
			// 		fmt.Printf("CAUSALITY VIOLATION: client %v received vector_clock %v from Client 4 when local clock is %v\n", id,msg5,vector_clock)
			// 	}
			// 	vector_clock = updateVectorClock(msg5,vector_clock,id)
        	// 	fmt.Printf("Client %v clock is now %v\n",id, vector_clock)
				// goSleep()
			case <- time.After(time.Second):
				// goSleep()
				vector_clock[id]= vector_clock[id]+1
				// TimeLagOnOneServer()
				goSleep()
				temp_copy_vector_clock := vector_clock
				// fmt.Printf("Client %v clock is now %v\n",id,vector_clock)
				if id == 1{
					random_select_client_int := rand.Intn(no_client)
					if random_select_client_int == id{
						random_select_client_int=2
					}
					go func(){
						fmt.Printf("SLEEPERINOSSSS: Client %v sends message %v to client %v\n",1,temp_copy_vector_clock,random_select_client_int)
						if random_select_client_int ==2{
							SleepT()
							fmt.Printf("SLEEPEDDDDDD. NOW CHECKKK: Client %v sends message %v to client %v\n",1,temp_copy_vector_clock,random_select_client_int)
						}else{
							print("IT SHOULD WORK AHHHH")
						}
						// fmt.Printf("Client %v sends message %v to client %v\n",1,vector_clock,random_select_client_int)
						clock_for_all[random_select_client_int][1] <- temp_copy_vector_clock
					}()
					
				}
				if id == 0 {
					random_select_client_int := rand.Intn(no_client)
					if random_select_client_int == id{
						random_select_client_int=1
					}
					go func(){
						fmt.Printf("Client %v sends message %v to client %v\n",0,temp_copy_vector_clock,random_select_client_int)
						clock_for_all[random_select_client_int][0] <- temp_copy_vector_clock
					}()
				}
				if id == 2  {
					random_select_client_int := rand.Intn(no_client)
					if random_select_client_int == id{
						random_select_client_int=0
					}
					go func(){
						fmt.Printf("Client %v sends message %v to client %v\n",2,temp_copy_vector_clock,random_select_client_int)
						clock_for_all[random_select_client_int][2] <- temp_copy_vector_clock

					}()
				}
				// if id == 3 {
				// 	random_select_client_int := rand.Intn(no_client)
				// 	if random_select_client_int == id{
				// 		random_select_client_int=0
				// 	}
				// 	fmt.Printf("Client %v sends message %v to client %v\n",3,vector_clock,random_select_client_int)
				// 	go func(){
				// 		clock_for_all[random_select_client_int][3] <- vector_clock
				// 	}()
				// }
				// if id == 4{
				// 	random_select_client_int := rand.Intn(no_client)
				// 	if random_select_client_int == id{
				// 		random_select_client_int=0
				// 	}
				// 	fmt.Printf("Client %v sends message %v to client %v\n",2,vector_clock,random_select_client_int)
				// 	go func(){
				// 	clock_for_all[random_select_client_int][4] <- vector_clock
				// 	}()
				// }
				TimeLagOnOneServer()
    	}
  	}
}



func testOnePointThree(){
	// we create an array of channels. These channels can send over a vector clock.
	var clock_for_all [no_client] [no_client]chan [no_client]int 
	for j:=0; j<no_client; j++  {
		var clock_for_one_client [no_client]chan [no_client]int 
		for i:=0; i<no_client; i++  {
			clock_for_one_client[i] = make(chan [no_client]int)
		}
		clock_for_all[j] = clock_for_one_client
	}
	go client(0, clock_for_all)
	go client(1, clock_for_all)
	go client(2, clock_for_all)
	// go client(3, clock_for_all)
	// go client(4, clock_for_all)
}


func main() {
	testOnePointThree()

	var input string
	fmt.Scanln(&input)
}
