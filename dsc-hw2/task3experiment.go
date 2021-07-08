package main5

import (
	"fmt"
	"math/rand"
	"time"
)

// const (
// no_client int = 5
// )

var GLOBAL_CLIENT_EXIT int = 0
var TIME_START int64 = 0

var NO_CLIENT_LIST [10]int = [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

// var NO_CLIENT_LIST []int = []int{1, 2, 3, 4}
var STORE_TIME_LIST map[int]int = make(map[int]int)
var CURRENT_NO_NODE int = 1 // counter for node number in experiments

/////////////////////////////////
/*helper functions*/ /////////////
/////////////////////////////////

type Node struct {
	id int
}
type Server struct {
	serverQueue []int
	lockFlag    bool
}

func goSleep() {
	amt := time.Duration(rand.Intn(1000))
	time.Sleep(time.Millisecond * amt)
	// fmt.Printf("server has slept for %v seconds\n", amt)
}
func SleepT() {
	amt := time.Duration(1000)
	time.Sleep(time.Millisecond * amt)
}
func retrieveMsgAndClientId(msg []string) (string, int) {
	for msgIndex, msgContent := range msg {
		if msgContent != "" {
			return msgContent, msgIndex
		}
	}
	return "", 0
}
func addToQueue(queue []int, value int) []int {
	queue = append(queue, value)
	return queue
}
func popQueue(queue []int) []int {
	return queue[1:]
}
func makeMessage(msg string, clientId int, no_client int) []string {
	var msgToServer []string = make([]string, no_client)
	for i := 0; i < no_client; i++ {
		msgToServer[i] = ""
	}
	msgToServer[clientId] = msg
	return msgToServer
}
func (n *Node) requestEnterCriticalSection(clock_client_to_server chan []string, no_client int) {
	msgToServer := makeMessage("request lock", n.id, no_client)
	go func(msgToServer []string) {
		clock_client_to_server <- msgToServer
	}(msgToServer)
}

//////////////////////////////////
/*client-server functions*/ ///////
//////////////////////////////////
func (n *Node) client(client_to_server chan []string, server_to_client []chan []string, no_client int) {
	// request once to enter critical section
	go n.requestEnterCriticalSection(client_to_server, no_client)
	for {
		// receive message from clock1 else check how many broadcast messages has client received
		select {
		case msg1 := <-server_to_client[n.id]:
			fmt.Printf("cleint %v receives message %v", n.id, msg1)
			messageType, getRequestClient := retrieveMsgAndClientId(msg1)
			if getRequestClient == n.id && messageType == "acquire lock" {
				//client has entered critical section
				fmt.Printf("CRITICAL SECTION: client %v has gotten the acquire lock.\n", n.id)
				SleepT()
				fmt.Printf("EXIT CRITICAL SECTION: client %v is exiting critical section.\n", n.id)
				msgToServer := makeMessage("release lock", n.id, no_client)
				go func(msgToServer []string) {
					client_to_server <- msgToServer
				}(msgToServer)
			}
		case <-time.After(time.Second):
			SleepT()
		}
	}
}

func (s *Server) server(client_to_server chan []string, server_to_client []chan []string, no_client int) {
	for {
		// receive message from clock1 or clock2 whichever is ready
		// if both clock1 and clock2 are ready, then break the tie randomly
		select {
		case msg1 := <-client_to_server: //retrieves the message to acquire lock
			fmt.Printf("server receives message %v\n", msg1)
			msgContent, clientId := retrieveMsgAndClientId(msg1)
			if msgContent == "" {
				fmt.Println("ERROR: client msg is blank.")
			} else {
				fmt.Printf("Server receiving %v from client %v\n", msgContent, clientId)
				if msgContent == "request lock" {
					//add clientid to queue
					s.serverQueue = addToQueue(s.serverQueue, clientId)
					// s.lockFlag = true
					if len(s.serverQueue) == 1 {
						go func(clientId int) {
							msgToServer := makeMessage("acquire lock", clientId, no_client)
							server_to_client[clientId] <- msgToServer
						}(clientId)
					}
				}
				if msgContent == "release lock" {
					//pop the first client in lock
					s.serverQueue = popQueue(s.serverQueue)
					fmt.Printf("what is in serverQueue?%v", s.serverQueue)
					GLOBAL_CLIENT_EXIT = GLOBAL_CLIENT_EXIT + 1
					if len(s.serverQueue) > 0 {
						go func(clientId int) {
							fmt.Printf("ermwhat is in serverQueue?%v", clientId)
							msgToServer := makeMessage("acquire lock", clientId, no_client)
							fmt.Printf("msgToServer?%v\n", msgToServer)
							server_to_client[clientId] <- msgToServer
						}(s.serverQueue[0])
					}
				}

			}
		case <-time.After(time.Second):
			SleepT()
		}
	}
}

func generateNode(no_client int) {
	TIME_START := time.Now().Unix()
	var nodeList []*Node
	var clock_client_to_server chan []string = make(chan []string)                // server should only read from one channel
	var clock_server_to_client []chan []string = make([]chan []string, no_client) // server can send out lock-acquire and client can send out lock-release msg seperately
	for i := 0; i < no_client; i++ {
		clock_server_to_client[i] = make(chan []string)
	}
	for i := 0; i < no_client; i++ {
		n := Node{
			id: i,
		}
		nodeList = append(nodeList, &n)
		go n.client(clock_client_to_server, clock_server_to_client, no_client)
	}
	s := Server{
		lockFlag: false,
	}
	go s.server(clock_client_to_server, clock_server_to_client, no_client)
	for {
		if GLOBAL_CLIENT_EXIT == no_client { //all nodes have exited critical section
			timeEnd := time.Now().Unix()
			timeElapsed := int(timeEnd - TIME_START)
			fmt.Printf("TIME ELAPSED: time elapsed for the experiment is : %v", timeElapsed)
			STORE_TIME_LIST[no_client] = timeElapsed
			CURRENT_NO_NODE = no_client + 1
			break
		}

	}

}

func main() {
	// generateNode()
	for i := 0; i < len(NO_CLIENT_LIST); i++ {
		for {
			// SleepT()
			// fmt.Println("STORE_TIME_LIST:", STORE_TIME_LIST)
			if NO_CLIENT_LIST[i] == CURRENT_NO_NODE {
				GLOBAL_CLIENT_EXIT = 0
				fmt.Printf("EXPERIMENT: NO_NODE %v\n", NO_CLIENT_LIST[i])
				generateNode(NO_CLIENT_LIST[i])
				break
			}
		}
	}
	for {
		SleepT()
		if _, ok := STORE_TIME_LIST[NO_CLIENT_LIST[len(NO_CLIENT_LIST)-1]]; ok {
			//all experiments complete
			fmt.Println("STORE_TIME_LIST:", STORE_TIME_LIST)
			break
		}
	}

	var input string
	fmt.Scanln(&input)
}
