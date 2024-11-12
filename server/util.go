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
	log.Printf("Node %d finished transition tasks. Now runs as Leader\n", n.Id)
	
	n.SendHeartbeats() // Initial heartbeat
	n.BeginStateTimer() // Then periodically will send out heartbeats
	// The following is purely for receiving messages
	// heartbeat and timeout is sent / handled automatically by RestartStateTimer 
	// for {
	// 	select {
	// 		case command := <-n.CommandChannel:
	// 			log.Printf("Leader Node %d received client command: %s\n", n.Id, command)
	// 			n.HandleClientCommand(command)
	// 		case <-n.QuitChannel:
	// 			log.Printf("Leader Node %d is stopping\n", n.Id)
	// 			return
	// 	}
	// }
	select {}
}

func (n *Node) RunAsFollower() {
	log.Printf("Node %d finished transition tasks. Now runs as Follower.\n", n.Id)

	go n.BeginStateTimer()
	// The following is purely for receiving messages
	// heartbeat and timeout is sent / handled automatically by RestartStateTimer 
	// for {
	// 	select {
	// 		case <-n.resetTimeoutChan:
	// 			// Received heartbeat, reset election timeout
	// 			// Remember to also set the n.LeaderID from heartbeat (consider case of new leader)
	// 			log.Printf("Node %d has received a heartbeat. Resetting timeout.\n", n.Id)
				
	// 		case <-time.After(n.TimeoutOrHeartbeatInterval):
	// 			log.Printf("Node %d heartbeat timeout. Becoming candidate.\n", n.Id)
	// 			n.BecomeCandidate()

	// 			return
	// 	}
	// }

	select {}
}

func (n *Node) RunAsCandidate() {
	log.Printf("Node %d finished transition tasks. Now runs as Candidate.\n", n.Id)

	go n.BeginStateTimer()
	// The following is purely for receiving messages
	// heartbeat and timeout is sent / handled automatically by RestartStateTimer 
	// for {
	// 	select {
	// 		case <-n.resetTimeoutChan:
	// 			// Received heartbeat. Then this node must concede
	// 			// Remember to also set the n.LeaderID from heartbeat (consider case of new leader)
	// 			log.Printf("Node %d has received a heartbeat. Conceding as candidate to become follower.\n", n.Id)

	// 			var dummyTerm int = 1 
	// 			n.BecomeFollower(dummyTerm)
				
	// 		case <-time.After(n.TimeoutOrHeartbeatInterval):
	// 			log.Printf("Node %d election timeout. Becoming candidate again.\n", n.Id)
	// 			n.BecomeCandidate()

	// 			return
	// 	}
	// }

	select {}
}


func (n *Node) StartRPCServer() {
	log.Printf("Node %d is attempting to register RPC\n", n.Id)

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

func GetFormatDuration(d time.Duration) string {
	seconds := d.Seconds()
	return fmt.Sprintf("%.5f seconds", seconds)
}