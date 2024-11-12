// The main file that creates the nodes and runs the nodes
package main

import (
	"flag"
	"fmt"
	"log"
	"raft/server"
	"time"
)

func main() {

	// Defining flags to control leader-failure simulation
	failLeader := flag.Bool("fail-leader", false, "Simulate leader failure after 5 seconds")
	flag.Parse()

	// Define channels for each node
	commandChannels := make(map[int]chan string)
	quitChannel := make(chan struct{}) // Channel to signal leader failure

	// Initialize 1 leader and 4 followers
	var nodes []*server.Node

	for i := 0; i < 5; i++ {
		node := server.NewNode(i, []int{0, 1, 2, 3, 4})
		node.SetTimeoutOrHeartbeatInterval() // Auto set timeout based on node's state
		node.QuitChannel = quitChannel // Channel to signal node to stop

		// Initialize fields for log replication
		node.Log = []server.LogEntry{} //Initial log as empty
		node.CurrentTerm = 1           // Initial term =1
		node.LeaderID = -1             // Initialize leaderID to -1 as no leader

		nodes = append(nodes, node)
		commandChannels[i] = make(chan string, 1) // Channel for client commands
		node.CommandChannel = commandChannels[i]
	}

	// Do not decide on leader node first. Everyone starts as follower

	// Start each node in its own goroutine with its specific channel
	for _, node := range nodes {

		go node.StartRPCServer()

		go func(n *server.Node) {
			n.RunAsFollower()
		}(node)
	}
	// Goroutine to print logs periodically
	if !*failLeader {

		go func() {
			ticker := time.NewTicker(5 * time.Second) // Adjust the interval as needed
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					log.Println("=== Snapshot of Logs from All Nodes ===")
					for _, node := range nodes {
						logEntries := node.GetLog()
						log.Printf("Node %d Log Entries:", node.Id)
						for i, entry := range logEntries {
							log.Printf("  Term: %d, Index: %d, Command: %v", entry.Term, i, entry.Command)
						}
					}
					log.Println("========================================")
				case <-quitChannel:
					return
				}
			}
		}()
	}
	// Simulate client commands sent directly to the leader's channel
	if !*failLeader {
		go func() {
			commandCounter :=1

			for {
			time.Sleep(3 * time.Second)
			log.Printf("========== Incoming Client command =========\n")

			command :=fmt.Sprintf("Client Command %d", commandCounter)
			commandChannels[0] <- command
			commandCounter ++
			}
		}()
	}

	// Simulate leader failure after 5 seconds (only once)
	// Added a flag to simulate leader failure. To run "go run main.go -fail-leader=true"
	if *failLeader {
		go func() {
			time.Sleep(5 * time.Second)
			log.Println("========= Simulating leader failure for Node 0 ============ \n")
			close(quitChannel) // Signal leader to stop sending heartbeats
		}()
	} else {
		log.Println("Leader failure simulation is disabled")
	}

	// Prevent main from exiting
	select {}
}
