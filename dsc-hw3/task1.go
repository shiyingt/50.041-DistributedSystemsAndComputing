package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	no_client = 5
	RW        = "READ/WRITE"
	RO        = "READ/ONLY"
	NILS      = "NIL"
)

// PAGECOUNT = 0

type CentralManagerRecord struct {
	copy_set   []int //contains the id of nodes holding a copy
	owner      int   //contains the id of owner node
	invalidate int
}
type CentralManager struct {
	data               map[int]CentralManagerRecord //map page_number to copy_set and owner
	directory_of_nodes []chan string                //number of channels will be number of nodes+1 and message format would be page_number-access
}
type Node struct {
	id                 int
	data               map[int]string // map page_number to access (r-w or r-o)
	directory_of_nodes []chan string  //number of channels will be number of nodes+1 and message format would be page_number-access
}

func createCentralManager(chan_nodes []chan string) *CentralManager {
	return &CentralManager{
		data:               make(map[int]CentralManagerRecord),
		directory_of_nodes: chan_nodes,
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
		owner: owner,
	}
}

/////////////////////
/*helper functions*/
////////////////////
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

////////////////////////

func (c *CentralManager) CM() {
	for {
		select {
		case msg1 := <-c.directory_of_nodes[no_client]:
			// fmt.Printf("server yes, msg: %v\n", msg1)
			message := msg1
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
					copysetClone := c.data[pageNumber].copy_set
					copysetClone = append(copysetClone, getRequestClient)
					dataClone := c.data[pageNumber]
					dataClone.copy_set = copysetClone
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
			// }
		}
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

//////////////
func generateNodes() {
	var nodeList []*Node
	var arr_channels []chan string = make([]chan string, no_client+1)
	for i := 0; i < no_client+1; i++ {
		arr_channels[i] = make(chan string)
	}
	for i := 0; i < no_client+1; i++ {
		n := createNode(i, arr_channels)
		nodeList = append(nodeList, n)
	}
	// for i := 0; i < no_client+1; i++ {
	// 	print(nodeList[i].id)
	// }
	server := createCentralManager(arr_channels)
	go server.CM()
	for i := 0; i < no_client; i++ {
		go nodeList[i].client()
	}
	// print("does it reach here?")
	nodeList[2].insertPage(1)
	SleepT()
	nodeList[0].readRequest(1)
	SleepT()
	nodeList[3].writeRequest(1)
	SleepT()
	nodeList[1].insertPage(2)
	SleepT()
	nodeList[0].readRequest(1)

}

func main() {
	generateNodes()
	var input string
	fmt.Scanln(&input)
}
