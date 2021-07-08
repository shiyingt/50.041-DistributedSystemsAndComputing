package main1

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	no_client                int = 5
	MESSAGE_PROPOGATION_TIME     = 1000
	MESSAGE_HANDLING_TIME        = 2000
)

var NO_CLIENT_LIST [10]int = [10]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
var STORE_TIME_LIST []int = make([]int, 10)
var GLOBAL_CLIENT_EXIT int = 0
var TIME_START int64 = 0

/////////////////////////////////
/*sleep helper functions*/ /////////////
/////////////////////////////////
func SleepRandom() {
	amt := time.Duration(rand.Intn(6000))
	time.Sleep(time.Millisecond * amt)
	// fmt.Printf("server has slept for %v seconds\n", amt)
}
func SleepT() {
	amt := time.Duration(1000)
	time.Sleep(time.Millisecond * amt)
}

/////////////////////////////////
/*helper functions*/ /////////////
/////////////////////////////////
func (n *Node1) broadcastRequest(broadcastMsg string) {
	for i := 0; i < len(n.directoryOfNodes); i++ {
		// fmt.Printf("client %v broadcasting. check don: %v\n", i, n.directoryOfNodes)
		if i != n.id {
			n.directoryOfNodes[i] <- broadcastMsg
		}
	}
}
func (n *Node1) broadcastReleaseMsg(broadcastMsg string) { //holds releaseCounter to check if experiment has terminated.
	for i := 0; i < len(n.directoryOfNodes); i++ {
		// fmt.Printf("client %v broadcasting. check don: %v\n", i, n.directoryOfNodes)
		// also broadcast release message to ownself to remove own timestamp-clientId entry + send accept message
		n.directoryOfNodes[i] <- broadcastMsg
	}
	GLOBAL_CLIENT_EXIT = GLOBAL_CLIENT_EXIT + 1
	if GLOBAL_CLIENT_EXIT == no_client { //all nodes have exited critical section
		timeEnd := time.Now().Unix()
		timeElapsed := timeEnd - TIME_START
		fmt.Printf("TIME ELAPSED: time elapsed for the experiment is : %v", timeElapsed)
	}

}
func returnMinimumValueInArray(arr []int) int {
	var minVal = arr[0]
	for i := 0; i < len(arr); i++ {
		if arr[i] < minVal {
			minVal = arr[i]
		}
	}
	return minVal
}
func retrieveClientIdWithSmallestTime(serverQueue []string) (int, int) {
	// disassemble the time - clientId pairing
	// get minimum amount of time from messages
	minTime, _ := strconv.Atoi(strings.Split(serverQueue[0], "-")[0])
	for i := 0; i < len(serverQueue); i++ {
		msg_split := strings.Split(serverQueue[i], "-")
		value, _ := strconv.Atoi(msg_split[0])
		if value < minTime {
			minTime = value
		}
	}
	//get all clients with the minimum timestamp and select the smallest ClientID
	var possibleClients []int
	for i := 0; i < len(serverQueue); i++ {
		msg_split := strings.Split(serverQueue[i], "-")
		value, _ := strconv.Atoi(msg_split[0])
		if value == minTime {
			pc, _ := strconv.Atoi(msg_split[1])
			possibleClients = append(possibleClients, pc)
		}
	}
	selectedClient := returnMinimumValueInArray(possibleClients)
	return minTime, selectedClient
}
func (n *Node1) checkIfAllReplyReceivedClientBroadcast() bool {
	for i := 0; i < len(n.repliesFromOtherMachines); i++ {
		// print(n.repliesFromOtherMachines[i])
		if n.repliesFromOtherMachines[i] == false {
			return false
		}
	}
	return true
}
func removeElementFromArray(arr []string, element string) []string {
	var index int = len(arr)
	for i := 0; i < len(arr); i++ {
		if element == arr[i] {
			index = i
		}
	}
	if index != len(arr) {
		// remove index of selected element from Array
		return append(arr[:index], arr[index+1:]...)
	}
	// else the element is not in array
	fmt.Println("NO ELEMENT IN ARRAY, element: %v, serverQUeue:%v", element, arr)
	return arr
}
func (n *Node1) requestEnterCriticalSection() {
	// set a random time so nodes start Critical Section at different times
	SleepRandom()
	// stamp request to enter with current time T
	timestamp := time.Now().Unix() //gets the unix time for timestamp
	makeReplies := make([]bool, no_client)
	n.repliesFromOtherMachines = makeReplies
	n.repliesFromOtherMachines[n.id] = true
	// fmt.Printf("ur checking %v", n.checkIfAllReplyReceivedClientBroadcast())
	// add request to queue
	var serverQueueMessage string = strconv.Itoa(int(timestamp)) + "-" + strconv.Itoa(n.id)
	n.serverQueue = append(n.serverQueue, serverQueueMessage)
	//broadcast request to other machines
	var broadcastMsg string = strconv.Itoa(int(timestamp)) + "-" + strconv.Itoa(n.id) + "-broadcast"
	print(broadcastMsg)
	go n.broadcastRequest(broadcastMsg)
	//wait till all replies from other machines are received
	for {

		SleepT()
		_, clientIdWithSmallestTime := retrieveClientIdWithSmallestTime(n.serverQueue)
		// smallestTime, clientIdWithSmallestTime := retrieveClientIdWithSmallestTime(n.serverQueue)
		fmt.Printf("currently in client %v queue is %v. what is the clientIdWithSmallestTime? %v. status of checkIfAllReplyReceivedClientBroadcast()?%v \n", n.id, n.serverQueue, clientIdWithSmallestTime, n.checkIfAllReplyReceivedClientBroadcast())
		if n.checkIfAllReplyReceivedClientBroadcast() == true && clientIdWithSmallestTime == n.id {
			//enter critical section
			fmt.Printf("CRITICAL SECTION: Client %v is in critical section\n", n.id)
			SleepT()
			fmt.Printf("EXIT CRITICAL SECTION:Client %v has exited critical section\n", n.id)
			//send release message to all machines
			var releaseMsg string = strconv.Itoa(int(timestamp)) + "-" + strconv.Itoa(n.id) + "-release"
			go n.broadcastReleaseMsg(releaseMsg)
			break
		}
	}
}

//////////////////////////////////
/*client-server functions*/ ///////
//////////////////////////////////
func (n *Node1) clientTask1() {
	// request once to enter critical section
	go n.requestEnterCriticalSection()
	for {
		// receive message from clock1 else check how many broadcast messages has client received
		select {
		case msg1 := <-n.directoryOfNodes[n.id]:
			fmt.Printf("client %v has received msg : %v\n", n.id, msg1)
			//add request to queue
			msg_split := strings.Split(msg1, "-") //msg{unix time, clientid, message type}
			if msg_split[2] == "accept" {
				// increase replies counter
				clientId, _ := strconv.Atoi(msg_split[1])
				n.repliesFromOtherMachines[clientId] = true
				fmt.Printf("client %v recieved R-ACCEPT from client %v\n", n.id, clientId)
			}
			if msg_split[2] == "broadcast" {
				//add request to queue
				timestamp, _ := strconv.Atoi(msg_split[0])
				clientId, _ := strconv.Atoi(msg_split[1])
				var serverQueueItem string = msg_split[0] + "-" + msg_split[1]
				n.serverQueue = append(n.serverQueue, serverQueueItem)
				minTime, nextClient := retrieveClientIdWithSmallestTime(n.serverQueue)
				fmt.Printf("nextClient: %v, n.serverQueue: %v \n", nextClient, n.serverQueue)
				if minTime == timestamp && nextClient == clientId {
					//reply to clientRequest
					var broadcastAccept string = strconv.Itoa(int(timestamp)) + "-" + strconv.Itoa(n.id) + "-accept"
					// fmt.Printf("ACCEPT:client %v sending accept message: %v, to client %v\n", n.id, broadcastAccept, clientId)
					go func(broadcastAccept string) {
						n.directoryOfNodes[nextClient] <- broadcastAccept
					}(broadcastAccept)
				}
			}
			if msg_split[2] == "release" {
				clientId, _ := strconv.Atoi(msg_split[1])
				var serverQueueItem string = msg_split[0] + "-" + msg_split[1]
				fmt.Printf("client %v recieved R-RELEASE from client %v\n", n.id, clientId)
				print(serverQueueItem)
				//pop head of queue
				n.serverQueue = removeElementFromArray(n.serverQueue, serverQueueItem)
				//send out next accept request
				if len(n.serverQueue) != 0 {
					minTime, nextClient := retrieveClientIdWithSmallestTime(n.serverQueue)
					fmt.Printf("nextClient: %v, n.serverQueue: %v \n", nextClient, n.serverQueue)
					if nextClient != n.id {
						//reply to clientRequest
						var broadcastAccept string = strconv.Itoa(int(minTime)) + "-" + strconv.Itoa(n.id) + "-accept"
						// fmt.Printf("ACCEPT:client %v sending accept message: %v, to client %v\n", n.id, broadcastAccept, clientId)
						go func(broadcastAccept string) {
							fmt.Printf("RELEASE-BROADCAST-ACCEPT")
							n.directoryOfNodes[nextClient] <- broadcastAccept
						}(broadcastAccept)
					}
				}
			}
			SleepT()
		case <-time.After(time.Second):
			SleepT()
		}
	}
}

type Node1 struct {
	id                       int
	serverQueue              []string // an array of "broadcast unix time - clientId" as received by the client. We implement it this way because there are some overlap in unix timings
	directoryOfNodes         map[int]chan string
	repliesFromOtherMachines []bool
}

func createNode(id int) *Node1 {
	node := Node1{
		id: id,
		// serverQueue: make([]string, 10),
	}
	return &node
}

func generateNode() {
	TIME_START = time.Now().Unix()
	var nodeList []*Node1
	var directoryOfNodes map[int]chan string = make(map[int]chan string, no_client)
	for i := 0; i < no_client; i++ {
		directoryOfNodes[i] = make(chan string)
	}
	for i := 0; i < no_client; i++ {
		n := createNode(i)
		n.directoryOfNodes = directoryOfNodes
		nodeList = append(nodeList, n)
		go n.clientTask1()
	}
}

func main() {
	generateNode()

	var input string
	fmt.Scanln(&input)
}
