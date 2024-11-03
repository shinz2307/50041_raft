// The main file that creates the nodes and runs the nodes
package main

import (
	"log"
	"time"
	"yourmodule/server"
)

func main() {
	heartbeatInterval := 400 * time.Millisecond
	electionTimeout := 5 * heartbeatInterval // Election timeout is 5x heartbeat interval

	// Define channels for each node
	heartbeatChannels := make(map[int]chan string)
	commandChannels := make(map[int]chan string)
	quitChannel := make(chan struct{}) // Channel to signal leader failure

	// Inidialise 1 leader 4 followers
	var nodes []*server.Node
	for i := 0; i < 5; i++ {
		node := server.NewNode(i, []int{0, 1, 2, 3, 4})
		node.ElectionTimeout = electionTimeout
		nodes = append(nodes, node)
		heartbeatChannels[i] = make(chan string, 1) // Channel for heartbeats
		commandChannels[i] = make(chan string, 1)   // Channel for client commands
	}

	// Leader node 0, Followers node 1-4
	nodes[0].State = server.Leader
	// Start each node in its own goroutine with its specific channel
	for _, node := range nodes {
		if node.State == server.Leader {
			go initialiseLeader(node, heartbeatChannels, commandChannels, heartbeatInterval, quitChannel)
		} else {
			go initialiseFollower(node, heartbeatChannels[node.Id], commandChannels[node.Id])
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
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Simulating leader failure for Node 0")
		close(quitChannel) // Signal leader to stop sending heartbeats
	}()

	// Prevent main from exiting
	select {}

}

func initialiseLeader(leader *server.Node, heartbeatChannels map[int]chan string, commandChannels map[int]chan string, heartbeatInterval time.Duration, quit chan struct{}) {
	ticker := time.NewTicker(400 * time.Millisecond) // Heartbeat interval
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			log.Printf("Leader %d: Stopping heartbeats due to simulated failure\n", leader.Id)
			return // Stop sending heartbeats

		case <-ticker.C:
			// Send heartbeats to all followers via their heartbeat channels
			for peerId, hbChan := range heartbeatChannels {
				if peerId != leader.Id {
					// AppendEntriesRPC()
					log.Printf("Leader %d: Sending heartbeat to Follower %d\n", leader.Id, peerId)
					hbChan <- "heartbeat"
				}
			}
		case command := <-commandChannels[leader.Id]:
			// Handle client command received on leader's command channel
			log.Printf("Leader %d: Received command to append: %s\n", leader.Id, command)
			leader.Log = append(leader.Log, server.LogEntry{Command: command})
		}
	}
}

func initialiseFollower(follower *server.Node, heartbeatChannel, commandChannel chan string) {
	timer := time.NewTimer(follower.ElectionTimeout) // Election timer
	defer timer.Stop()

	for {
		select {
		case heartbeat, ok := <-heartbeatChannel:
			if !ok {
				// Heartbeat channel closed, meaning the leader has failed
				log.Printf("Follower %d: Detected leader failure\n", follower.Id)
				follower.State = server.Candidate
				//startElection(follower, heartbeatChannel, commandChannel)
				return
			}
			// if else logic in the event of empty heartbeat and heartbeat with
			log.Printf("Follower %d: Received heartbeat %s, resetting election timer\n", follower.Id, heartbeat)
			// ReceiveAppendEntriesRPC()
			if !timer.Stop() {
				<-timer.C // Drain the timer channel
			}
			timer.Reset(follower.ElectionTimeout)

		// Process command if needed (followers might queue commands until a leader is available)
		case command := <-commandChannel:
			log.Printf("Follower %d: Received client command: %s\n", follower.Id, command)
			// To forward to leader, commandChannels append?

		// No heartbeat received in time, transition to candidate
		case <-timer.C:
			log.Printf("Follower %d: Election timeout, becoming candidate\n", follower.Id)
			follower.State = server.Candidate
			//startElection(follower, heartbeatChannel, commandChannel)
			return
		}
	}
}
