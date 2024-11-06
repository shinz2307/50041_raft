//go:build exclude
// +build exclude

package server

import (
	"fmt"
	"log"
	"net/rpc"
)

// HeartbeatArgs represents the arguments for the Heartbeat RPC method
type HeartbeatArgs struct {
	// Include any necessary fields here
}

// HeartbeatReply represents the response from the Heartbeat RPC method
type HeartbeatReply struct {
	// Include any necessary fields here
}

// Function will run on leader
func (n *Node) SendHeartbeats() {
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue // dont send heartbeat to myself
		}
		go func(id int) {
			args := &HeartbeatArgs{}
			var reply HeartbeatReply
			err := n.CallHeartbeatRPC(id, args, &reply)
			if err != nil {
				log.Printf("Leader Node %d failed to send heartbeat to Node %d: %v\n", n.Id, id, err)
			}
		}(peerID)
	}
}

// Function will run on the leader
func (n *Node) CallHeartbeatRPC(peerID int, args *HeartbeatArgs, reply *HeartbeatReply) error {
	address := fmt.Sprintf("localhost:%d", 8000+peerID) // Construct the target address
	log.Printf("Heartbeat sending to %v\n", address)
	client, err := rpc.Dial("tcp", address) // Establish an RPC connection to the follower node
	if err != nil {                         // Handle connection error
		return err
	}
	defer client.Close() // Ensure the connection is closed after the RPC call

	// Use the unique service name to call the Heartbeat method on the target node
	serviceName := fmt.Sprintf("Node%d.Heartbeat", peerID)
	return client.Call(serviceName, args, reply) // Make the RPC call to the follower's Heartbeat method
}

// Function will run on a follower or candidate
func (n *Node) Heartbeat(args *HeartbeatArgs, reply *HeartbeatReply) error {
	// Handle the heartbeat reception logic
	// For example, reset the election timeout
	select {
	case n.resetTimeoutChan <- struct{}{}:
		// Election timeout reset successfully
		log.Printf("Node %d received heartbeat\n", n.Id)
	default:
		// Channel was full; no action needed
	}
	// Optionally, you can populate the reply
	return nil // No error occurred
}
