// server/heartbeatRPC.go

package server

import (
	"log"
	"fmt"
	"net/rpc"
)

// AppendEntriesRequest represents the request structure for the AppendEntries RPC.
type AppendEntriesRequest struct {
	Term         int         // Leader's term
	LeaderID     int         // Leader's ID
	PrevLogIndex int         // Index of log entry immediately preceding new ones
	PrevLogTerm  int         // Term of PrevLogIndex entry
	Entries      []LogEntry  // Log entries to store (empty for heartbeat)
	LeaderCommit int         // Leader’s commit index
}

// AppendEntriesResponse represents the response structure for the AppendEntries RPC.
type AppendEntriesResponse struct {
	Term    int  // Current term for leader to update itself
	Success bool // True if follower contained entry matching PrevLogIndex and PrevLogTerm
}
// AppendEntries handles incoming AppendEntries RPC requests on followers.
func (n *Node) AppendEntries(args *AppendEntriesRequest, reply *AppendEntriesResponse) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Case: If follower's term > leader's term, return false
	if args.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}

	// Follower updates the term to match leader and becomes a follower (if it was a candidate)
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.State = Follower
		n.LeaderID = args.LeaderID
	}

	// Append any new entries not already in the log
	for i, entry := range args.Entries {
		index := args.PrevLogIndex + 1 + i
		if index < len(n.Log) {
			if n.Log[index].Term != entry.Term {
				// Conflict detected, delete the existing entry and all that follow it
				n.Log = n.Log[:index]
				n.Log = append(n.Log, entry)
			}
		} else {
			n.Log = append(n.Log, entry)
		}
	}

	reply.Term = n.CurrentTerm
	reply.Success = true
	log.Printf("Node %d appended entries from Leader %d", n.Id, args.LeaderID)
	return nil
}

// SendAppendEntries sends AppendEntries RPC to a specific follower.
func (n *Node) SendAppendEntries(peerID int, entries []LogEntry) {
	n.mu.Lock()
	prevLogIndex := len(n.Log) - 1
	prevLogTerm := 0
	if prevLogIndex >= 0 {
		prevLogTerm = n.Log[prevLogIndex].Term
	}
	args := &AppendEntriesRequest{
		Term:         n.CurrentTerm,
		LeaderID:     n.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
	}
	n.mu.Unlock()

	var reply AppendEntriesResponse
	err := n.CallAppendEntriesRPC(peerID, args, &reply)
	if err != nil {
		log.Printf("Leader Node %d failed to send AppendEntries to Node %d: %v\n", n.Id, peerID, err)
		return
	}

	if reply.Success {
		log.Printf("Leader Node %d successfully replicated entries to Node %d", n.Id, peerID)
		// You can add additional logic here if needed
	} else {
		if reply.Term > n.CurrentTerm {
			n.mu.Lock()
			n.CurrentTerm = reply.Term
			n.State = Follower
			n.LeaderID = -1
			n.mu.Unlock()
		} else {
			// Handle log inconsistency, possibly retry or adjust nextIndex
			log.Printf("Leader Node %d detected log inconsistency with Node %d", n.Id, peerID)
		}
	}
}

// CallAppendEntriesRPC performs the actual RPC call to AppendEntries on the follower.
func (n *Node) CallAppendEntriesRPC(peerID int, args *AppendEntriesRequest, reply *AppendEntriesResponse) error {
	address := fmt.Sprintf("localhost:%d", 8000+peerID)
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer client.Close()

	serviceName := fmt.Sprintf("Node%d", peerID)
	return client.Call(serviceName+".AppendEntries", args, reply)
}

// HandleClientCommand processes a client command received by the leader.
func (n *Node) HandleClientCommand(command string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Append the command to the leader's log
	entry := LogEntry{
		Command: command,
		Term:    n.CurrentTerm,
	}
	n.Log = append(n.Log, entry)
	log.Printf("Leader Node %d appended command: %s", n.Id, command)

	// Send AppendEntries to all followers with the new entry
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue // Skip sending to self
		}
		go n.SendAppendEntries(peerID, []LogEntry{entry})
	}
}