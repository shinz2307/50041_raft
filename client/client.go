package main

import (
	"log"
	"net/rpc"
	"time"
)

func main() {
	// Add a startup delay to allow nodes to initialise
	startupDelay := 10 * time.Second
	log.Printf("Waiting %v for nodes to start up...", startupDelay)
	time.Sleep(startupDelay)

	nodes := []string{"app0:8080", "app1:8080", "app2:8080", "app3:8080", "app4:8080"}
	var leader string
	var client *rpc.Client
	var err error

	// Retry mechanism to find the leader
	for _, address := range nodes {
		client, err = rpc.Dial("tcp", address)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", address, err)
			continue
		}

		// Check if this node is the leader
		var isLeader bool
		err = client.Call("SingleNode.IsLeader", &struct{}{}, &isLeader)
		if err != nil {
			log.Printf("Failed to call IsLeader on %s: %v", address, err)
			client.Close()
			continue
		}

		if isLeader {
			leader = address
			log.Printf("Found leader: %s", leader)
			break
		}
		client.Close() // Close connection if not the leader
	}

	if leader == "" {
		log.Fatalf("Could not find the leader after checking all nodes")
	}

	// Connect to the leader
	client, err = rpc.Dial("tcp", leader)
	if err != nil {
		log.Fatalf("Failed to connect to the leader: %v", err)
	}
	defer client.Close()

	// Send a command to the leader
	command := "Sample Command"
	var reply bool
	err = client.Call("SingleNode.HandleClientCommand", &command, &reply)
	if err != nil {
		log.Fatalf("Failed to send command: %v", err)
	}

	log.Printf("Command sent successfully, reply: %v", reply)
}
