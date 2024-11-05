// The main file that creates the nodes and runs the nodes
package main

import (
	//"log"
	"raft/server"
	"time"
)

func main() {
	heartbeatInterval := 2000 * time.Millisecond
	electionTimeout := 5 * heartbeatInterval // Election timeout is 5x heartbeat interval

	// Define channels for each node
	commandChannels := make(map[int]chan string)
	quitChannel := make(chan struct{}) // Channel to signal leader failure

	// Initialize 1 leader and 4 followers
	var nodes []*server.Node

	for i := 0; i < 5; i++ {
		node := server.NewNode(i, []int{0, 1, 2, 3, 4})
		node.ElectionTimeout = electionTimeout
		node.HeartbeatInterval = heartbeatInterval
		node.QuitChannel = quitChannel // Channel to signal node to stop
		
		// Initialize fields for log replication 
		node.Log = []server.LogEntry{} //Initial log as empty 
		node.CurrentTerm = 1 // Initial term =1 
		node.LeaderID = -1 // Initialize leaderID to -1 as no leader

		nodes = append(nodes, node)
		commandChannels[i] = make(chan string, 1) // Channel for client commands
		node.CommandChannel = commandChannels[i]
	}

	// Leader node 0, Followers node 1-4
	nodes[0].State = server.Leader
	nodes[0].LeaderID = 0 //Set initial leaderID to its own

	// Start each node in its own goroutine with its specific channel
	for _, node := range nodes {
		if node.State == server.Leader {
			// Initialize as leader
			go func(n *server.Node) {
				n.RunAsLeader()
			}(node)
		} else {
			// Initialize as follower
			go func(n *server.Node) {
				n.RunAsFollower()
			}(node)
		}
	}

	// Simulate client commands sent directly to the leader's channel
	go func() {
		for {
			commandChannels[0] <- "Client Command"
			time.Sleep(2 * time.Second)
		}
	}()

	// Simulate leader failure after 5 seconds (only once)
	// Commenting out to test for appendEntries
	/**go func() {
		time.Sleep(5 * time.Second)
		log.Println("Simulating leader failure for Node 0")
		close(quitChannel) // Signal leader to stop sending heartbeats
	}()
		*/
	// Prevent main from exiting
	select {}
}