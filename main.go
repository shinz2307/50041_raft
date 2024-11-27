// The main file that creates the nodes and runs the nodes
package main

import (
	"flag"
	"fmt"
	"log"
	"raft/server"
	"raft/shared"
	"time"
)

func main() {

	// Defining flags to control leader-failure simulation
	failLeader := flag.Bool("fail-leader", false, "Simulate leader failure after 5 seconds")
	failClient := flag.Bool("fail-client", false, "Simulate Follower Node failure after 5 seconds")
	// NewLeader := flag.Bool("new-leader", false, "Simulate New Leader elected and new different Index")
	higherTerm := flag.Bool("higher-term", false, "Simulate Different logs on Node 1")
	flag.Parse()

	// Define channels for each node
	quitChannel := make(chan struct{}) // Channel to signal leader failure

	var nodes []*server.Node

	commandChannel := make(chan string, 1)

	// Initialise 1 new leader and 4 followers with already existing logs

	// Initialize 1 leader and 4 followers
	for i := 0; i < 5; i++ {
		node := server.NewNode(i, []int{0, 1, 2, 3, 4})
		node.SetTimeoutOrHeartbeatInterval() // Auto set timeout based on node's state
		node.QuitChannel = quitChannel       // Channel to signal node to stop
		node.ElectionTimeout = electionTimeout
		node.HeartbeatInterval = heartbeatInterval
		node.QuitChannel = quitChannel // Channel to signal node to stop

		// Initialize fields for log replication
		node.Log = []server.LogEntry{} //Initial log as empty
		node.CurrentTerm = 1           // Initial term =1
		node.LeaderID = -1             // Initialize leaderID to -1 as no leader

		nodes = append(nodes, node)
		node.CommandChannel = commandChannel
	}

	if *shared.NewLeader {
		// Leader[0]: [1, 1, 4, 4, 5, 6]
		// 1: [1, 1, 3, 3]
		// 2: [1, 1, 2]
		// 3: [1, 1, 4, 4, 5, 6, 6]
		// 4: [1, 1, 4, 4, 4]
		nodes[0].CurrentTerm = 6
		nodes[1].CurrentTerm = 3
		nodes[2].CurrentTerm = 2
		nodes[3].CurrentTerm = 6
		nodes[4].CurrentTerm = 4
		nodes[0].NextIndex = map[int]int{
			1: 7,
			2: 7,
			3: 7,
			4: 7,
		}

		for i := 0; i < 2; i++ {
			entry := server.LogEntry{
				Command: "command 1",
				Term:    1,
			}
			for _, node := range nodes {
				node.Log = append(node.Log, entry)
			}
		}
		// New leader, Node 3 and 4 has two of command 4
		for i := 0; i < 2; i++ {
			entry := server.LogEntry{
				Command: "command 4",
				Term:    4,
			}
			nodes[0].Log = append(nodes[0].Log, entry)
			nodes[3].Log = append(nodes[3].Log, entry)
			nodes[4].Log = append(nodes[4].Log, entry)
		}
		// Node 4 has an extra command 4 [1, 1, 4, 4, 4]
		entry := server.LogEntry{
			Command: "command 4",
			Term:    4,
		}

		nodes[4].Log = append(nodes[4].Log, entry)

		// Only Node 2 has command 2 [1, 1, 2]
		entry = server.LogEntry{
			Command: "command 2",
			Term:    2,
		}
		nodes[2].Log = append(nodes[2].Log, entry)

		// Node 1 has two of command 3 [1, 1, 3, 3]
		for i := 0; i < 2; i++ {
			entry := server.LogEntry{
				Command: "command 3",
				Term:    3,
			}
			nodes[1].Log = append(nodes[1].Log, entry)
		}
		// Leader and Node 3 has command 5
		entry = server.LogEntry{
			Command: "command 5",
			Term:    5,
		}
		nodes[0].Log = append(nodes[0].Log, entry)
		nodes[3].Log = append(nodes[3].Log, entry)
		// Leader has one of command 6 but Node 3 has two of command 6
		entry = server.LogEntry{
			Command: "command 6",
			Term:    6,
		}
		nodes[0].Log = append(nodes[0].Log, entry)
		nodes[3].Log = append(nodes[3].Log, entry)
		nodes[3].Log = append(nodes[3].Log, entry)

	} else {
		for _, node := range nodes {
			// Initialize fields for log replication
			node.Log = []server.LogEntry{} //Initial log as empty
			node.CurrentTerm = 1           // Initial term =1
			node.LeaderID = -1             // Initialize leaderID to -1 as no leader
		}
	}

	// Do not decide on leader node first. Everyone starts as follower

	// Leader node 0, Followers node 1-4
	nodes[0].State = server.Leader
	nodes[0].LeaderID = 0 //Set initial leaderID to its own

	// Start each node in its own goroutine with its specific channel
	for _, node := range nodes {

		go node.StartRPCServer()
		go func(n *server.Node) {
			n.RunAsFollower()
		}(node)
	}
	if *higherTerm {
		entry := server.LogEntry{
			Command: "command",
			Term:    2,
		}
		log.Println("========= Simulating Different Logs for Node 1 ============ \n")
		nodes[1].Log = append(nodes[1].Log, entry)
		nodes[1].CurrentTerm = 2
		nodes[1].CommitIndex = 1
	}
	if *higherTerm {
		entry := server.LogEntry{
			Command: "command",
			Term:    2,
		}
		log.Println("========= Simulating Different Logs for Node 1 ============ \n")
		nodes[1].Log = append(nodes[1].Log, entry)
		nodes[1].CurrentTerm = 2
		nodes[1].CommitIndex = 1
	}
	// Goroutine to print logs periodically
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Adjust the interval as needed
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("=== Snapshot of Logs from All Nodes ===")
				for _, node := range nodes {
					logEntries := node.GetLog()
					log.Printf("Node %d Current Term: %d Commit Index: %d", node.Id, node.CurrentTerm, node.CommitIndex)
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

	// Simulate client commands sent directly to the leader's channel
	if !*shared.NewLeader {
		go func() {
			commandCounter := 1

			for {
				time.Sleep(5 * time.Second)
				log.Printf("========== Incoming Client command =========\n")

				command := fmt.Sprintf("Client Command %d", commandCounter)
				commandChannels[0] <- command
				commandCounter++
			}
		}()
	}

	// log.Printf("========== Incoming Client command =========\n")

	// command := fmt.Sprintf("Client Command %d")
	// commandChannels[0] <- command

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

	// Client fails and restarts later

	if *failClient {
		go func() {
			time.Sleep(15 * time.Second)
			log.Println("========= Simulating Node 1 Failure ============ \n")
			nodes[1].Failed = true

			// Restart the client after 5 seconds
			time.Sleep(10 * time.Second)
			nodes[1].Failed = false
			log.Println("========= Node 1 Recovered ========== \n")
		}()
	}

	// Prevent main from exiting
	select {}
}
