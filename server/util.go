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
	fmt.Printf("Node %d is starting up as Leader\n", n.Id)

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
			fmt.Printf("Leader Node %d is stopping\n", n.Id)
			return
		}
	}
}

func (n *Node) RunAsFollower() {
	fmt.Printf("Node %d is running as Follower\n", n.Id)

	go n.StartRPCServer()

	for {
		electionTimeout := n.ElectionTimeout

		select {
		case <-n.resetTimeoutChan:
			// Received heartbeat, reset election timeout
		case <-time.After(electionTimeout):
			fmt.Printf("Node %d election timeout. Becoming candidate.\n", n.Id)
			// can add election logic here
			return
		}
	}
}

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
				fmt.Printf("Leader Node %d failed to send heartbeat to Node %d: %v\n", n.Id, id, err)
			}
		}(peerID)
	}
}

func (n *Node) StartRPCServer() {
	fmt.Printf("Node %d is attempting to register\n", n.Id)

	if err := rpc.Register(n); err != nil {
		if err.Error() == "rpc: service already defined: Node" {
			log.Printf("Node %d has already been registered\n", n.Id)
			return
		}
		log.Fatalf("Failed to register Node %d for RPC: %v\n", n.Id, err)
	}

	// generate port number attach listener to address
	address := fmt.Sprintf("localhost:%d", 8000+n.Id)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Node %d failed to listen: %v\n", n.Id, err)
	}
	defer listener.Close()
	rpc.Accept(listener)
}

func (n *Node) CallHeartbeatRPC(peerID int, args *HeartbeatArgs, reply *HeartbeatReply) error {
	address := fmt.Sprintf("localhost:%d", 8000+peerID) // Construct the target address
	client, err := rpc.Dial("tcp", address)             // Establish an RPC connection to the follower node
	if err != nil {                                     // Handle connection error
		return err
	}
	defer client.Close() // Ensure the connection is closed after the RPC call

	return client.Call("Node.Heartbeat", args, reply) // Make the RPC call to the follower's Heartbeat method
}

func (n *Node) Heartbeat(args *HeartbeatArgs, reply *HeartbeatReply) error {
	// Handle the heartbeat reception logic
	// For example, reset the election timeout
	select {
	case n.resetTimeoutChan <- struct{}{}:
		// Election timeout reset successfully
	default:
		// Channel was full; no action needed
	}
	// Optionally, you can populate the reply
	return nil // No error occurred
}
