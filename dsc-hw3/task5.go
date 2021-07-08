package main5

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	replication_client       = 3
	no_client                = 5
	RW                       = "READ/WRITE"
	RO                       = "READ/ONLY"
	NILS                     = "NIL"
	MESSAGE_PROPOGATION_TIME = 2
	MESSAGE_HANDLING_TIME    = 3
)

var REJECT_ELECTION_TIME_LIMIT int = MESSAGE_PROPOGATION_TIME*2 + MESSAGE_HANDLING_TIME

var TIME_START time.Time
var GLOBAL_READ_COUNT = 0
var READ_TIMINGS map[int]time.Duration = make(map[int]time.Duration)

// PAGECOUNT = 0

type CentralManagerRecord struct {
	copy_set   []int //contains the id of nodes holding a copy
	owner      int   //contains the id of owner node
	invalidate int
}

func changeCentralManagerRecordIntoMessage(data map[int]CentralManagerRecord) string {
	// page_number:copyset_1!copyset2*page_owner*invalidate_int,
	var entry string = ""
	for k, v := range data {
		var copysetString string = ""
		for _, copy := range v.copy_set {
			copysetString = copysetString + strconv.Itoa(copy) + "!"
		}
		if copysetString != "" {
			copysetString = copysetString[:len(copysetString)-1]
		}
		var record string = strconv.Itoa(k) + ":" + copysetString + "*" + strconv.Itoa(v.owner) + "*" + strconv.Itoa(v.invalidate)
		entry = entry + record + ","
	}
	return entry[:len(entry)-1]
}
func changeMessageIntoCentralManagerRecord(data string) map[int]CentralManagerRecord {
	// fmt.Printf("changeMessageIntoCentralManagerRecord: %v", data)
	var mapCMR map[int]CentralManagerRecord = make(map[int]CentralManagerRecord)
	CMRstring := strings.Split(data, ",")
	for _, record := range CMRstring {
		//decode string
		recordSplit := strings.Split(record, ":")        // recordSplit[0] is page_number
		CMRstructs := strings.Split(recordSplit[1], "*") //<copysetString,owner,invalidate_int>
		//create new record
		page_number, _ := strconv.Atoi(recordSplit[0])
		owner, _ := strconv.Atoi(CMRstructs[1])
		invalidate, _ := strconv.Atoi(CMRstructs[2])
		var newRecord CentralManagerRecord = createCentralManagerRecord(owner)
		newRecord.invalidate = invalidate

		if len(CMRstructs[0]) != 0 { //copysetList is 0
			copysetList := strings.Split(CMRstructs[0], "!")
			var copysetIntList []int = make([]int, len(copysetList))
			for i := 0; i < len(copysetList); i++ {
				intCopy, _ := strconv.Atoi(copysetList[i])
				copysetIntList[i] = intCopy
			}
			newRecord.copy_set = copysetIntList
		}

		mapCMR[page_number] = newRecord
	}
	return mapCMR
}

type CentralManager struct {
	id                           int
	coordinator_id               int
	alive                        bool //sets if the status of central manager is alive or dead
	election_stage               string
	rejection_flag               bool
	election_directory_of_nodes  map[int]chan string //channels for all central managers
	control_panel_for_alive_node map[int]chan bool   // maps ids of node to channel, use this channel to send instructions to killNode.
	coordinator_is_alive         bool

	data               map[int]CentralManagerRecord //map page_number to copy_set and owner
	directory_of_nodes []chan string                //channels for  number of nodes+1(the primary) and message format would be page_number-access
}
type Node struct {
	id                 int
	data               map[int]string // map page_number to access (r-w or r-o)
	directory_of_nodes []chan string  //number of channels will be number of nodes+1 and message format would be page_number-access
}

func createCentralManager(chan_nodes []chan string, id int, election_directory_of_nodes map[int]chan string, control_panel map[int]chan bool) *CentralManager {
	return &CentralManager{
		data:                         make(map[int]CentralManagerRecord),
		directory_of_nodes:           chan_nodes,
		id:                           id,
		election_stage:               "",
		rejection_flag:               false,
		coordinator_id:               -1,
		alive:                        true,
		coordinator_is_alive:         false,
		election_directory_of_nodes:  election_directory_of_nodes,
		control_panel_for_alive_node: control_panel,
	}
}
func createNode(id int, chan_nodes []chan string) *Node {
	return &Node{
		id:                 id,
		data:               make(map[int]string),
		directory_of_nodes: chan_nodes,
	}
}
func createCentralManagerRecord(owner int) CentralManagerRecord {
	return CentralManagerRecord{
		owner:      owner,
		invalidate: 0,
	}
}

/////////////////////
/*helper functions*/
////////////////////
func find(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
func makeMessage(msg string, clientId int) []string {
	var msgToServer []string = make([]string, no_client+1)
	for i := 0; i < no_client+1; i++ {
		msgToServer[i] = ""
	}
	msgToServer[clientId] = msg
	return msgToServer
}
func SleepT() {
	amt := time.Duration(1000)
	time.Sleep(time.Millisecond * amt)
}
func SleepTimeout() {
	time.Sleep(time.Duration(2*MESSAGE_PROPOGATION_TIME+MESSAGE_HANDLING_TIME) * time.Second)
}

func (c *CentralManager) returnIfHigherId(incoming_node_id int) bool {
	if c.id > incoming_node_id {
		return true
	} else {
		return false
	}
}

// func retrieveMsgAndClientId(msg []string) (string, int) {
// 	for msgIndex, msgContent := range msg {
// 		if msgContent != "" {
// 			return msgContent, msgIndex
// 		}
// 	}
// 	return "", 0
// }

///////////////////
/*node functions*/
//////////////////
//types of actions that a node can request from central manager: store, read, write
func (n *Node) insertPage(page_number int) {
	//add page to data property in Node w/ r/w access in node
	n.data[page_number] = RW
	//send the page to central manager, where client manager's channel is of index[no_client]
	//message should be of the form: <page_number,action,clientSenderId>
	var storePage string = strconv.Itoa(page_number) + "-store-" + strconv.Itoa(n.id)
	go func(msgStorePage string) {
		// print("it reaches here")
		n.directory_of_nodes[no_client] <- msgStorePage
	}(storePage)
}
func (n *Node) readRequest(page_number int) {
	//check own data first
	var checkIfKeyExists bool = false
	if _, ok := n.data[page_number]; ok {
		if n.data[page_number] != NILS {
			checkIfKeyExists = true
			fmt.Printf("there exists a local cache for page number %v at Client %v", page_number, n.id)
		} else {
			fmt.Printf("the page number in the local cache has been invalidated.\n")
		}
	}
	if checkIfKeyExists == false {
		//message should be of the form: <page_number,action,clientSenderId>
		var readRequest string = strconv.Itoa(page_number) + "-read-" + strconv.Itoa(n.id)
		go func(msg string) {
			n.directory_of_nodes[no_client] <- msg
		}(readRequest)
	}

}
func (n *Node) writeRequest(page_number int) {
	//message should be of the form: <page_number,action,clientSenderId>
	var readRequest string = strconv.Itoa(page_number) + "-write-" + strconv.Itoa(n.id)
	go func(msg string) {
		n.directory_of_nodes[no_client] <- msg
	}(readRequest)
}
func (c *CentralManager) killCentralManager() {
	c.alive = false
	c.control_panel_for_alive_node[c.id] <- false
}
func (c *CentralManager) broadcastDataCM(kill_channel chan bool) {
	for {
		select {
		case msg2 := <-kill_channel:
			if msg2 == false {
				//kill broadcast data
				fmt.Printf("broadcasting for CM %v was killed\n", c.id)
				return
			}
		case <-time.After(3 * time.Second): //broadcast to all replica nodes the data in CM
			if len(c.data) != 0 {
				for i := 0; i < len(c.election_directory_of_nodes); i++ {
					if i != c.id {
						go func(i int) {
							msg := changeCentralManagerRecordIntoMessage(c.data)
							var m string = "CMR&" + msg
							c.election_directory_of_nodes[i] <- m
						}(i)
					}
				}
			}

		}
	}
}

////////////////////////
//only the primary replica is supposed to run this function
func (c *CentralManager) CM() {
	fmt.Printf("Central Manager %v reporting for work", c.id)
	//periodically send out data to replica CM
	var channel_replica chan bool = make(chan bool)
	go c.broadcastDataCM(channel_replica)
	for {
		select {
		case msg2 := <-c.control_panel_for_alive_node[c.id]:
			if msg2 == false {
				channel_replica <- false
				fmt.Printf("Central Manager with id %v was killed :(", c.id)
				//quits
				return
			}
		case msg1 := <-c.directory_of_nodes[no_client]:
			// fmt.Printf("server yes, msg: %v\n", msg1)
			message := msg1
			// if c.coordinator_id==c.id{	// it is the primary replica
			// message, checkIfClientRecipientIsNode := retrieveMsgAndClientId(msg1)
			// if checkIfClientRecipientIsNode == no_client { //message is intended for recipient
			fmt.Printf("server received message: %v\n", message)
			msgSplit := strings.Split(message, "-") //{page_number,actions} where actions:<store,read,write,copy>
			pageNumber, _ := strconv.Atoi(msgSplit[0])
			getRequestClient, _ := strconv.Atoi(msgSplit[2])
			if msgSplit[1] == "store" { //ask central manager to store info of page in central manager's data
				// create a record and add it to central manager
				record := createCentralManagerRecord(getRequestClient)
				c.data[pageNumber] = record
			}
			if msgSplit[1] == "read" { //ask central manager to redirect read request to clinet
				//fetch owner of page
				if _, ok := c.data[pageNumber]; ok {
					// if owner of page found, then forward the request
					var readRequest string = msgSplit[0] + "-read-" + strconv.Itoa(no_client) + "-" + strconv.Itoa(getRequestClient) //message:<page_number,action,clientSenderId,requestClientId>
					// msg := makeMessage(readRequest, c.data[pageNumber].owner)
					go func(msg string) {
						c.directory_of_nodes[c.data[pageNumber].owner] <- msg
					}(readRequest)
				}
			}
			if msgSplit[1] == "copy" { //adds to the copyset in the central manager
				//fetch owner of page
				if _, ok := c.data[pageNumber]; ok {
					// if owner of page found, update the copyset in central manager
					// fmt.Printf("Copysetclone: %v", c.data[pageNumber].copy_set)
					if !find(c.data[pageNumber].copy_set, getRequestClient) {
						copysetClone := c.data[pageNumber].copy_set
						copysetClone = append(copysetClone, getRequestClient)
						dataClone := c.data[pageNumber]
						dataClone.copy_set = copysetClone
						// fmt.Printf("dataClone copyset: %v", dataClone.copy_set)
						c.data[pageNumber] = dataClone
						// the first time we detect a change in the copysetchange the read-write access of owner to read-only
						if len(c.data[pageNumber].copy_set) == 1 {
							//message:<page_number,action,clientSenderId>
							var readOnlyRequest string = msgSplit[0] + "-readOnly-" + strconv.Itoa(no_client)
							// msg := makeMessage(readOnlyRequest, c.data[pageNumber].owner)
							go func(msg string) {
								c.directory_of_nodes[c.data[pageNumber].owner] <- msg
							}(readOnlyRequest)
						}
					}
				}
				//EXPERIMENT-READ: we assume that upon reaching here the read request for a page is completed.
				GLOBAL_READ_COUNT = GLOBAL_READ_COUNT + 1
				fmt.Printf("GLOBAL_READ_COUNT%v\n", GLOBAL_READ_COUNT)
				timeDiff := time.Since(TIME_START)
				READ_TIMINGS[GLOBAL_READ_COUNT] = timeDiff

			}
			if msgSplit[1] == "write" {
				//send message from the central message to invalidate all the copies of page
				//message:<page_number,action,clientSenderId,requestClientId>
				var invalidateMsg = msgSplit[0] + "-invalidate-" + strconv.Itoa(no_client) + "-" + strconv.Itoa(getRequestClient)
				if _, ok := c.data[pageNumber]; ok {
					for i := 0; i < len(c.data[pageNumber].copy_set); i++ {
						go func(clientId int) {
							// msg := makeMessage(invalidateMsg, clientId)
							c.directory_of_nodes[clientId] <- invalidateMsg
						}(c.data[pageNumber].copy_set[i])
					}
				}
			}
			if msgSplit[1] == "invalidateConfirmation" {
				//receive invalidateConfirmationMessage from nodes
				if _, ok := c.data[pageNumber]; ok {
					copysetClone := c.data[pageNumber].invalidate
					copysetClone = copysetClone + 1
					dataClone := c.data[pageNumber]
					dataClone.invalidate = copysetClone
					c.data[pageNumber] = dataClone
				}
				//if all confirmation from nodes in copyset are received, we can now write forward to node
				if c.data[pageNumber].invalidate == len(c.data[pageNumber].copy_set) {
					//invalidate copyset
					var newArr []int
					dataClone := c.data[pageNumber]
					dataClone.copy_set = newArr
					c.data[pageNumber] = dataClone
					//write forward to original owner of page'
					//message:<page_number,action,clientSenderId,requestClientId>
					var writeMsg = msgSplit[0] + "-write-" + strconv.Itoa(no_client) + "-" + msgSplit[3]
					// msg := makeMessage(writeMsg, c.data[pageNumber].owner)
					go func(msg string) {
						c.directory_of_nodes[c.data[pageNumber].owner] <- msg
					}(writeMsg)
				}
			}
			if msgSplit[1] == "writeConfirmation" {
				if _, ok := c.data[pageNumber]; ok {
					//change owner of page
					//make invalidation count 0 again
					dataClone := c.data[pageNumber]
					dataClone.invalidate = 0
					dataClone.owner = getRequestClient
					c.data[pageNumber] = dataClone
				}
			}

		}
		// }
	}
}

func (n *Node) client() {
	for {
		select {
		case msg1 := <-n.directory_of_nodes[n.id]:
			// fmt.Printf("client yes, msg: %v\n", msg1)
			message := msg1
			// message, checkIfClientRecipientIsNode := retrieveMsgAndClientId(msg1) //requestClient is always server
			// if checkIfClientRecipientIsNode == n.id {                             // check if message is meant for node
			fmt.Printf("client %v received message: %v\n", n.id, message)
			msgSplit := strings.Split(message, "-") //{page_number,actions,clientSenderId,requestClient} where actions:<store,read,write>
			pageNumber, _ := strconv.Atoi(msgSplit[0])
			// getRequestClient, _ := strconv.Atoi(msgSplit[2])
			if msgSplit[1] == "read" { // client receives "read" forwarded by server
				//message:<page_number,action,clientSenderId>
				var readRequest string = msgSplit[0] + "-store-" + strconv.Itoa(n.id)
				requestClientId, _ := strconv.Atoi(msgSplit[3])
				// msg := makeMessage(readRequest, requestClientId)
				go func(msg string) {
					n.directory_of_nodes[requestClientId] <- msg
				}(readRequest)
			}
			if msgSplit[1] == "store" { //client receives "store" from another client which is a response from a previouly issued "read" request to the central manager
				n.data[pageNumber] = RO //store read-only for page
				//send copyset to central manager
				//message:<page_number,action,clientSenderId>
				var storeRequest string = msgSplit[0] + "-copy-" + strconv.Itoa(n.id)
				// msg := makeMessage(storeRequest, no_client)
				go func(msg string) {
					n.directory_of_nodes[no_client] <- msg
				}(storeRequest)
			}
			if msgSplit[1] == "invalidate" { //client receives "invalidate" from server which is a response to "write" request to server
				n.data[pageNumber] = NILS
				//message:<page_number,action,clientSenderId,requestClientId>
				var invalidateConfirmationMsg string = msgSplit[0] + "-invalidateConfirmation-" + strconv.Itoa(n.id) + "-" + msgSplit[3]
				// msg := makeMessage(invalidateConfirmationMsg, no_client)
				go func(msg string) {
					n.directory_of_nodes[no_client] <- msg
				}(invalidateConfirmationMsg)
			}
			if msgSplit[1] == "write" { //client received write forward from server, of type <page_number,action,requestClient>
				//client sends page to client who request write access
				//message:<page_number,action,clientSenderId>
				var writeAccessMsg string = msgSplit[0] + "-writeAccess-" + strconv.Itoa(n.id)
				requestClientId, _ := strconv.Atoi(msgSplit[3])
				// msg := makeMessage(writeAccessMsg, requestClientId)
				go func(msg string) {
					n.directory_of_nodes[requestClientId] <- msg
				}(writeAccessMsg)
				//client invalidates its own access
				n.data[pageNumber] = NILS
			}
			if msgSplit[1] == "writeAccess" {
				//update access in own data
				n.data[pageNumber] = RW
				//write confirmation to central manager
				//message:<page_number,action,clientSenderId>
				var writeConfirmationMsg string = msgSplit[0] + "-writeConfirmation-" + strconv.Itoa(n.id)
				// msg := makeMessage(writeConfirmationMsg, no_client)
				go func(msg string) {
					n.directory_of_nodes[no_client] <- msg
				}(writeConfirmationMsg)
			}
			if msgSplit[1] == "readOnly" {
				n.data[pageNumber] = RO
			}
			// }
		}
	}

}

/////////////////////////////////////////////////////
/////////////////////////////////////////////////////
/////////////////////////////////////////////////////
//////////////////
/*Election*/ ////////
//////////////////
func (c *CentralManager) callElection() {
	// STAGE1 : ELECTION
	fmt.Printf("election starting for Central Manager %v\n", c.id)
	if !(c.election_stage == "election") {
		c.election_stage = "election"
		c.rejection_flag = false
		var electionMsg string = "elect: Central Manager %v " + strconv.Itoa(c.id)
		// send election message only to nodes with higher known ids
		if c.alive == true {
			for k := 0; k < len(c.election_directory_of_nodes); k++ {
				if !c.returnIfHigherId(k) {
					if !(c.id == k) { //no need to send elect message to ownself
						go func(c *CentralManager, k int, electionMsg string) {
							c.election_directory_of_nodes[k] <- electionMsg
						}(c, k, electionMsg)
					}
				}
			}
		}
		// STAGE 2: ANSWER
		//wait for certain time for reject message before declaring that we are coordinator
		// waitRejectMessage := time.Now().Unix()
		// for {
		// 	if !(int(time.Now().Unix()-waitRejectMessage) < REJECT_ELECTION_TIME_LIMIT*10 && c.alive == false) {
		// 		break
		// 	}
		// }
		if c.alive == true {
			SleepTimeout()
			SleepTimeout()
			// SleepTimeout()
			// SleepTimeout()
		}
		fmt.Printf("CM %v rejection flag:%v\n", c.id, c.rejection_flag)
		// STAGE 3: BROADCAST
		if c.rejection_flag == false && c.alive == true {
			c.coordinator_id = c.id
			for k, _ := range c.directory_of_nodes {
				if !(c.id == k) { //no need to send broadcasr message to ownself
					// send election message only to nodes with higher known ids
					var broadcastMsg = "broadcast msg: the new coordinator is Node " + strconv.Itoa(c.id)
					// fmt.Printf("CLIENT %v SENDING BROADCAST MSG TO%v\n", n.id, k)
					go func(k int) {
						c.election_directory_of_nodes[k] <- broadcastMsg
					}(k)
				}
			}
			fmt.Printf("starting CM for coordinator %v", c.id)
			go c.CM()
		}
		c.election_stage = ""
	}

}

func (c *CentralManager) bootstrap_election_communication() {
	for {
		select {
		case msg1 := <-c.election_directory_of_nodes[c.id]:
			// fmt.Printf("[]CLIENT %v HAS RECEIVED MSG%V\n",n.id,msg1)
			if c.alive == false {
				//don't do any work, because node is supposed to be dead
			} else {
				// fmt.Printf("[]CENTRAL MANAGER %v HAS RECEIVED MSG%v\n", c.id, msg1)
				// make sure node is alive
				//check if is reject message
				if strings.Contains(msg1, "reject") {
					c.rejection_flag = true
					fmt.Printf("CM %v set rejection flag to true\n", c.id)
				}
				//check if its election_msg
				if strings.Contains(msg1, "elect") {
					//check who started the election
					election_initator_id := msg1[len(msg1)-1:]
					election_initator_id_int, _ := strconv.Atoi(election_initator_id)
					fmt.Printf("CM %v received msg: %v , and election_initiator_id %v\n", c.id, msg1, election_initator_id_int)
					if c.returnIfHigherId(election_initator_id_int) {
						// send rejection message
						go func(election_initator_id_int int) {
							fmt.Printf("Client %v sending a reject election message to client %v\n", c.id, election_initator_id_int)
							c.election_directory_of_nodes[election_initator_id_int] <- "reject"
						}(election_initator_id_int)
						// start election if not in election
						if c.election_stage == "" {
							go func() {
								c.callElection()
							}()
						}
					}
				}
				// check if its broadcast_msg
				if strings.Contains(msg1, "broadcast") {
					//if this node was previously the coordinator, stop the CM() and broadcast
					if c.coordinator_id == c.id {
						fmt.Printf("stopping operations for central manager %v", c.id)
						go func() {
							c.control_panel_for_alive_node[c.id] <- false
						}()
					}
					//check who started the election
					coordinator_id := msg1[len(msg1)-1:]
					coordinator_id_int, _ := strconv.Atoi(coordinator_id)
					c.coordinator_id = coordinator_id_int
					fmt.Printf("Central Manager %v has set the coordinator id to %v\n", c.id, c.coordinator_id)

				}
				// node is coordinator. It responds to pingCoordinatorAlive message.
				if strings.Contains(msg1, "are you alive?") && c.coordinator_id == c.id {
					request_id := msg1[len(msg1)-1:]
					request_id_int, _ := strconv.Atoi(request_id)
					fmt.Printf("Coordinator %v has received pingCoordinatorAlive message from CM %v\n", c.id, request_id_int)
					go func() {
						c.election_directory_of_nodes[request_id_int] <- "ya im alive"
					}()
				}
				// node receives response from coordinator that is is alive.
				if strings.Contains(msg1, "ya im alive") {
					c.coordinator_is_alive = true
					fmt.Printf("CM %v has received pingCoordinatorAlive message from Coordinator\n", c.id)
				}
				if strings.Contains(msg1, "CMR") { // this is a message from primary replica to update the secondary Central manager replicas
					fmt.Printf("CM %v has a message from primary replica to update the secondary Central manager replicas\n", c.id)
					msgString := strings.Split(msg1, "&")
					cmr := changeMessageIntoCentralManagerRecord(msgString[1])
					c.data = cmr
					fmt.Printf("CM %v 's struct for data now: %v\n", c.id, c.data)
				}
			}
		case msg2 := <-c.control_panel_for_alive_node[c.id]:
			// check instructions on state of node - is it set to be alive or dead?
			if msg2 == false {
				c.alive = false
			} else {
				c.alive = true
			}
		case <-time.After(3 * time.Second):
			// at every regular interval, check if coordinator is still alive. else start new election
			if c.coordinator_id != -1 { //coordinator has been set before
				if c.id != c.coordinator_id { //coordinator node does not need to send pings to coordinators node
					if c.alive == true { //make sure node is alive
						go func(c *CentralManager) {
							var pingCoordinatorAlive = "are you alive? I am Node " + strconv.Itoa(c.id)
							c.election_directory_of_nodes[c.coordinator_id] <- pingCoordinatorAlive
							SleepTimeout()
							SleepTimeout()
							SleepTimeout()
							//coordinator is alive
							if c.coordinator_id != -1 { //check if previous coordinator has been set and callElection was already called.
								if c.coordinator_is_alive == false {
									fmt.Printf("Coordinator %v did not respond before timeout. Client %v restarting election\n", c.coordinator_id, c.id)
									c.coordinator_id = -1
									if c.election_stage == "" {
										go c.callElection()
									}
									// print("yes coordinator ping exists properly")
								} else {
									fmt.Printf("CM %v knows that coordinator is still alive\n", c.id)
									SleepTimeout()
									c.coordinator_is_alive = false
								}
							}
						}(c)
					}
				}
			}
			SleepTimeout()

		}
	}
}

/////////////////////////////////////////////////////
/////////////////////////////////////////////////////
/////////////////////////////////////////////////////
func generateNodes() {
	var nodeList []*Node
	var CMList []*CentralManager
	var arr_channels []chan string = make([]chan string, no_client+1)
	var election_channel map[int]chan string = make(map[int]chan string)
	var channel_kill_node map[int]chan bool = make(map[int]chan bool)
	for i := 0; i < no_client+1; i++ {
		arr_channels[i] = make(chan string)
	}

	for i := 0; i < no_client+1; i++ {
		n := createNode(i, arr_channels)
		nodeList = append(nodeList, n)
	}
	for i := 0; i < replication_client; i++ {
		election_channel[i] = make(chan string)
		channel_kill_node[i] = make(chan bool)
	}
	// print(election_channel)
	// print("hello")
	// print(channel_kill_node)
	for i := 0; i < replication_client; i++ {
		server := createCentralManager(arr_channels, i, election_channel, channel_kill_node)
		// server.election_directory_of_nodes = election_channel
		// server.control_panel_for_alive_node = channel_kill_node
		CMList = append(CMList, server)

	}
	// for i := 0; i < no_client+1; i++ {
	// 	print(nodeList[i].id)
	// }
	// for i := 0; i < replication_client; i++ {
	// 	print(CMList[i].id)
	// }
	// print(CMList[0].election_directory_of_nodes)
	// print("hi")
	// server := createCentralManager(arr_channels)
	for i := 0; i < replication_client; i++ {
		// go CMList[i].CM()
		go CMList[i].bootstrap_election_communication()
	}

	for i := 0; i < no_client; i++ {
		go nodeList[i].client()
	}
	CMList[replication_client-1].callElection()
	randWriteClient := rand.Intn(no_client - 1)
	randPage := rand.Intn(100)
	nodeList[randWriteClient].insertPage(randPage)
	SleepT()
	SleepT()
	// nodeList[1].readRequest(randPage)
	TIME_START = time.Now() // get time taken for 30 read requests
	for i := 0; i < 30; i++ {
		randReadClient := rand.Intn(no_client - 1)
		nodeList[randReadClient].readRequest(randPage)
	}

}

func main() {
	generateNodes()
	for {
		if len(READ_TIMINGS) == 20 {
			//is done
			fmt.Printf("[FT]READ TIMINGS :%v\n", READ_TIMINGS)
			return
		}
	}
	var input string
	fmt.Scanln(&input)
}
