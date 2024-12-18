// Helper functions for timeouts, randomisation, etc.
package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	// "raft/shared"
	"time"
)

func (n *Node) RunAsLeader() {
	log.Printf("Node %d finished transition tasks. Now runs as Leader\n", n.Id)

	// Initialise NextIndex map
	n.mu.Lock()
	n.NextIndex = make(map[int]int)
	n.mu.Unlock()

	for _, peerID := range n.Peers {
		if peerID != n.Id {
			// Initialising NextIndex for each peer to the length of the leader's log
			n.mu.Lock()
			n.NextIndex[peerID] = len(n.Log)
			n.mu.Unlock()
		}
	}
	// log.Printf("Node %d initialised NextIndex: %v", n.Id, n.NextIndex)

	n.SendHeartbeats()  // Initial heartbeat
	n.BeginStateTimer() // Then periodically will send out heartbeats
	// The following is purely for receiving messages
	// heartbeat and timeout is sent / handled automatically by RestartStateTimer

	// if !*shared.NewLeader { // Added
	// go n.InitializeNextIndex(n.Peers)

	log.Printf("Leader's log: %v", n.Log)
	if len(n.Log) != 0 {
		for _, peerID := range n.Peers {
			if peerID == n.Id {
				continue // Skip sending to self
			}
			go n.SendAppendEntries(peerID, n.Log)
		}
	}
	for {
		log.Printf("Node %d is now listening for client commands\n", n.Id)
		select {
		case command := <-n.CommandChannel: // Receive a command from the client
			log.Printf("Leader Node %d received client command: %s\n", n.Id, command)

			// Call HandleClientCommand as an RPC-like internal function
			reply := false
			err := n.HandleClientRead(&command, &reply)
			if err != nil {
				log.Printf("Error handling client command on Node %d: %v\n", n.Id, err)
			} else if reply {
				log.Printf("Leader Node %d successfully processed client command: %s\n", n.Id, command)
			}
		case <-n.QuitChannel:
			log.Printf("Leader Node %d is stopping\n", n.Id)
			return
		}
	}
}

func (n *Node) RunAsFollower() {
	log.Printf("Node %d finished transition tasks. Now runs as Follower.\n", n.Id)

	n.BeginStateTimer()
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

	n.BeginStateTimer()
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

func (n *Node) StartSingleRPCServer() {
	// log.Printf("Node %d is attempting to register RPC\n", n.Id)

	// Since we are using the same struct, we need to create new server to register the RPC
	server := rpc.NewServer() // Create a new server for each node
	if err := server.RegisterName("SingleNode", n); err != nil {
		log.Fatalf("Failed to register for RPC: %v\n", err)
	}

	// hard code internal port number
	address := "0.0.0.0:8080"
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}
	defer listener.Close()

	log.Printf("RPC server is listening on %s...\n", address)

	// Accept incoming connections
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
