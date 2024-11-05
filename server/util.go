// Helper functions for timeouts, randomisation, etc.
package server

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
)

func (n *Node) RunAsLeader() {
	log.Printf("Node %d is starting up as Leader\n", n.Id)

	go n.StartRPCServer()

	// maybe add delay?
	time.Sleep(2 * time.Second)

	ticker := time.NewTicker(n.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		// Every tick, send heartbeat
		case <-ticker.C:
			n.SendHeartbeats()
		case <-n.QuitChannel:
			log.Printf("Leader Node %d is stopping\n", n.Id)
			return
		}
	}
}

func (n *Node) RunAsFollower() {
	log.Printf("Node %d is running as Follower\n", n.Id)

	go n.StartRPCServer()

	for {
		electionTimeout := n.ElectionTimeout

		select {
		case <-n.resetTimeoutChan:
			// Received heartbeat, reset election timeout
			// Remember to also set the n.LeaderID from heartbeat (consider case of new leader)
		case <-time.After(electionTimeout):
			log.Printf("Node %d election timeout. Becoming candidate.\n", n.Id)
			n.StartElection()

			return
		}
	}
}


func (n *Node) StartRPCServer() {
	log.Printf("Node %d is attempting to register\n", n.Id)

	// Since we are using the same struct, we need to create new server to register the RPC 
	server := rpc.NewServer() // Create a new server for each node
	nodeServiceName := fmt.Sprintf("Node%d", n.Id)
	if err := server.RegisterName(nodeServiceName, n); err != nil {
		log.Fatalf("Failed to register %s for RPC: %v\n", nodeServiceName, err)
	}

	// generate port number attach listener to address
	address := fmt.Sprintf("localhost:%d", 8000+n.Id)
	log.Printf("Node %d trying to listen on %v...\n", n.Id, address)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Node %d failed to listen: %v\n", n.Id, err)
	} 

	defer listener.Close()
	log.Printf("Node %d listening on %v...\n", n.Id, address)
	for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }
        go server.ServeConn(conn)
    }
}