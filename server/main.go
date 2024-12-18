package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {

	// Read NODE_ID from environment variable
	nodeIDEnv := os.Getenv("NODE_ID")
	if nodeIDEnv == "" {
		log.Fatal("NODE_ID environment variable is not set!")
	}

	var nodeID int
	_, err := fmt.Sscanf(nodeIDEnv, "%d", &nodeID)
	if err != nil {
		log.Fatalf("Invalid NODE_ID: %v", err)
	}

	// Read PEERS from environment variable
	peersEnv := os.Getenv("PEERS")
	if peersEnv == "" {
		log.Fatal("PEERS environment variable is not set!")
	}

	// Parse PEERS into a slice of integers
	peers := parsePeers(peersEnv)

	// Initialize a single node based on NODE_ID
	node := initializeSingleNode(nodeID, peers)

	log.Printf("Node %d initialized with peers: %v", nodeID, peers)
	go node.StartSingleRPCServer()
	go node.RunAsFollower()

	// Prevent main from exiting

	// time.Sleep(10 * time.Second)
	// if node.State == Leader {
	// 	node.Failed = true
	// }
	select {}
}

// parsePeers parses a comma-separated string of peer IDs into a slice of integers
func parsePeers(peersEnv string) []int {
	peerStrings := strings.Split(peersEnv, ",")
	peers := make([]int, len(peerStrings))

	for i, peer := range peerStrings {
		peerID, err := strconv.Atoi(peer)
		if err != nil {
			log.Fatalf("Invalid peer ID in PEERS: %v", err)
		}
		peers[i] = peerID
	}

	return peers
}

// initializeSingleNode initializes a single node based on its ID.
func initializeSingleNode(nodeID int, peers []int) *Node {
	node := NewNode(nodeID, peers) // Peers are all nodes in the cluster
	node.SetTimeoutOrHeartbeatInterval()

	node.Log = []LogEntry{}
	node.CurrentTerm = 0
	node.LeaderID = -1
	return node
}
