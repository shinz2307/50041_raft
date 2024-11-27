// We can just overwrite main to structure our testing

package main

import (
	"log"
	"raft/server"
	"sync"
)

var numberOfNodes = 2

func main() {

    log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	quitChannel := make(chan struct{})

	var nodes []*server.Node

	commandChannel := make(chan string, 1)

	for i := 0; i < numberOfNodes; i++ {
		node := server.NewNode(i, []int{0, 1})
		node.SetTimeoutOrHeartbeatInterval()
		node.QuitChannel = quitChannel

		node.Log = []server.LogEntry{}
		node.CurrentTerm = 1
		node.LeaderID = -1

		nodes = append(nodes, node)
		node.CommandChannel = commandChannel
		go node.StartRPCServer()

	}

	var setupWait sync.WaitGroup
	setupWait.Add(numberOfNodes)

	for i := 0; i < numberOfNodes; i++ {

		if i == 0 {
			go func(){
				defer setupWait.Done()
				nodes[i].BecomeCandidate()
				log.Printf("wait()")
			}()
		} else {
			go func(term int){
				defer setupWait.Done()
				nodes[i].BecomeFollower(term)
				log.Printf("wait()")
			}(0)
		}
		
	}
	setupWait.Wait()
	log.Printf("Wait completed")

	var quitWait sync.WaitGroup
	quitWait.Add(numberOfNodes)
	log.Printf("Beginning to read the states")
	go func(nodes []*server.Node){
		if nodes[0].VoteCount == 1 {
			log.Printf("Node 0 received a vote from node 1!")
		}


		quitWait.Done()
	}(nodes)

	quitWait.Wait()
	log.Printf("Wait completed\n")
	close(quitChannel)
	select {}
}
