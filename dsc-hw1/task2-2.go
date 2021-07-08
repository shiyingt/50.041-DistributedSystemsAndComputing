///////////////
/*node struct*/
///////////////
package main
import (
	"time"
    "strings"
	"fmt"
	"strconv"
)
const MESSAGE_PROPOGATION_TIME int = 1
const MESSAGE_HANDLING_TIME int = 1

type Node struct{
	id int
	directory_of_nodes map[int] chan string //maps id of node to channel, use this channel to send string to other nodes
	stage string
	rejection_flag bool
	coordinator_id int
	alive bool // state of node
	control_panel_for_alive_node map[int] chan bool // maps ids of node to channel, use this channel to send instructions to killNode. 
	coordinator_is_alive bool
}

////////////////////
/*helper functions*/
////////////////////
func SleepT(){
	amt := time.Duration(1000)
    time.Sleep(time.Millisecond * amt)
}
func SleepTimeout(){
	time.Sleep(time.Duration(2*MESSAGE_PROPOGATION_TIME+MESSAGE_HANDLING_TIME)*time.Second)
}
func (n *Node) returnIfHigherId(incoming_node_id int) bool{
	if n.id>incoming_node_id{
		return true
	}else {
		return false
	}
}

//////////////////
/*Election*/////////
//////////////////
func (n *Node) callElection(){
	// STAGE1 : ELECTION
	fmt.Printf("election starting for client %v\n", n.id)
	if !(n.stage =="election"){
		n.stage = "election"
		n.rejection_flag = false
		var electionMsg string = "elect me: Node "+strconv.Itoa(n.id)
		// send election message only to nodes with higher known ids
		if n.alive==true{
			for k, _ := range n.directory_of_nodes{
				if !n.returnIfHigherId(k){
					if !(n.id==k){	//no need to send elect message to ownself
						// fmt.Printf("client %v sending election message to %v\n", n.id,k)
						// fmt.Printf("K IS :%v\n",k)
						go func(n *Node, k int, electionMsg string ){
							n.directory_of_nodes[k] <- electionMsg
						}(n, k, electionMsg)
					}
				}
			}
		}
		// STAGE 2: ANSWER
		//wait for certain time for reject message before declaring that we are coordinator
		if n.alive==true{
			SleepTimeout()
			fmt.Printf("client %v rejection flag:%v\n",n.id,n.rejection_flag)
		}
		// STAGE 3: BROADCAST 
		if n.rejection_flag == false && n.alive==true{
			n.coordinator_id=n.id
			for k, _ := range n.directory_of_nodes{
				if !(n.id==k){	//no need to send broadcasr message to ownself
					// send election message only to nodes with higher known ids
					var broadcastMsg = "broadcast msg: the new coordinator is Node "+strconv.Itoa(n.id)
					// fmt.Printf("CLIENT %v SENDING BROADCAST MSG TO%v\n", n.id, k)
					n.directory_of_nodes[k] <- broadcastMsg
				}
			}
		}
		n.stage=""
	}
	
}

func (n *Node) bootstrap_client_communication(){
	for {
		select{
			case msg1 := <- n.directory_of_nodes[n.id]:
				// fmt.Printf("[]CLIENT %v HAS RECEIVED MSG%V\n",n.id,msg1)
				if n.alive == false{
					//don't do any work, because node is supposed to be dead
				}else{
					// fmt.Printf("[]CLIENT %v HAS RECEIVED MSG%v\n",n.id,msg1)
					// make sure node is alive
					//check if is reject message
					if strings.Contains(msg1,"reject"){
						n.rejection_flag = true
						fmt.Printf("client %v set rejection flag to true\n", n.id)
					}
					//check if its election_msg
					if strings.Contains(msg1,"elect"){
						//check who started the election
						election_initator_id := msg1[len(msg1)-1:]
						election_initator_id_int,_ := strconv.Atoi(election_initator_id)
						fmt.Printf("client %v received msg: %v , and election_initiator_id %v\n", n.id, msg1, election_initator_id_int )
						if n.returnIfHigherId(election_initator_id_int){
							// send rejection message
							go func (){
								fmt.Printf("Client %v sending a reject election message to client %v\n", n.id, election_initator_id_int)
								n.directory_of_nodes[election_initator_id_int]<-"reject"
							}()
							// start election
							go func (){
								n.callElection()
							}()
						}
					}
					// check if its broadcast_msg
					if strings.Contains(msg1,"broadcast"){
						//check who started the election
						coordinator_id := msg1[len(msg1)-1:]
						coordinator_id_int,_ := strconv.Atoi(coordinator_id)
						n.coordinator_id = coordinator_id_int
						fmt.Printf("Client %v has set the coordinator id to %v\n",n.id,n.coordinator_id)
					}
					// node is coordinator. It responds to pingCoordinatorAlive message.
					if strings.Contains(msg1,"are you alive?") && n.coordinator_id == n.id{
						request_id := msg1[len(msg1)-1:]
						request_id_int,_ := strconv.Atoi(request_id)
						fmt.Printf("Coordinator %v has received pingCoordinatorAlive message from Client %v\n",n.id,request_id_int)
						n.directory_of_nodes[request_id_int]<-"ya im alive"
					}
					// node receives response from coordinator that is is alive.
					if strings.Contains(msg1,"ya im alive") {
						n.coordinator_is_alive=true
						fmt.Printf("Client %v has received pingCoordinatorAlive message from Coordinator\n",n.id)
					}
				}
			case msg2 := <- n.control_panel_for_alive_node[n.id]:
				// check instructions on state of node - is it set to be alive or dead?
				if msg2 == false{
					n.alive = false
				}else{
					n.alive = true
				}
			case <- time.After(time.Second):
				// at every regular interval, check if coordinator is still alive. else start new election
				if !(n.coordinator_id == -1){		//coordinator has been set before
					if (n.id!= n.coordinator_id){	//coordinator node does not need to send pings to coordinators node 
						if n.alive == true{	//make sure node is alive 
							go func(){
								var pingCoordinatorAlive = "are you alive? I am Node "+strconv.Itoa(n.id)
								n.directory_of_nodes[n.coordinator_id]<-pingCoordinatorAlive
								SleepTimeout()
								//coordinator is alive
								if n.coordinator_is_alive==false{
									fmt.Printf("Coordinator did not respond before timeout. Client %v restarting election\n",n.id)
									n.coordinator_id =-1
									n.callElection()
								}else{
									fmt.Printf("Client %v knows that coordinator is still alive\n",n.id)
									SleepTimeout()
									n.coordinator_is_alive = false
								}
							}()
						}
					}
				}
				SleepTimeout()


		}
	}
}

func createNode(id int) *Node{
	node := Node{
		id: id,
		directory_of_nodes: make(map[int] chan string),
		stage: "",
		rejection_flag: false,
		coordinator_id:-1,
		alive: true,
		coordinator_is_alive: false,
	}
	return &node
}


func generateNode(c string, no_c int){
	var no_client int = no_c
	var all_node []*Node
	var channel_all map[int] chan string = make(map[int] chan string)
	var channel_kill_node map[int] chan bool = make(map[int] chan bool)
	for i:=0; i<no_client; i++  {
		channel_all[i]= make(chan string)
		channel_kill_node[i]= make(chan bool)
	}
	for i:=0;i<no_client;i++{
		all_node = append(all_node,createNode(i))
		all_node[i].directory_of_nodes = channel_all
		all_node[i].control_panel_for_alive_node = channel_kill_node
		fmt.Printf("here lies a node %v\n",all_node[i].id)
		go all_node[i].bootstrap_client_communication()
	}
	//best case
	if c == "best case"{
		all_node[no_client-1].callElection()
	}
	//worst case
	if c == "worst case"{
		all_node[0].callElection()
	}
	// HIHI HERE WE INDUCE THE NODE FAILURES
	time.Sleep(3000)
	// kill the coordinator node 
	fmt.Println("Sent message to KILL coordinator node")
	channel_kill_node[no_client-1] <-false
}
func bestCase(no_client int){
	generateNode("best case",no_client)
}
func worstCase(no_client int){
	generateNode("worst case",no_client)
}
func main(){
	//you can change the number of clients
	const no_client = 5
	// run best case. Run seperately from worst case.
	bestCase(no_client)
	//uncomment to run worst case. comment out the best case before running the file.
	// worstCase(no_client)
	var input string
	fmt.Scanln(&input)
}