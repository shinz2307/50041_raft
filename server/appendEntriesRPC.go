// server/heartbeatRPC.go

package main

import (
	"fmt"
	"log"
	"net/rpc"
	"time"
)

var successChan = make(chan bool, 100)

// AppendEntriesRequest represents the request structure for the AppendEntries RPC.
type AppendEntriesRequest struct {
	Term         int        // Leader's term
	LeaderID     int        // Leader's ID
	PrevLogIndex int        // Index of log entry immediately preceding new ones
	PrevLogTerm  int        // Term of PrevLogIndex entry
	Entries      []LogEntry // Log entries to store (empty for heartbeat)
	LeaderCommit int        // Leader’s commit index
}

// AppendEntriesResponse represents the response structure for the AppendEntries RPC.
type AppendEntriesResponse struct {
	Term       int  // Current term for leader to update itself
	Success    bool // True if follower contained entry matching PrevLogIndex and PrevLogTerm
	NodeFailed bool
}

// AppendEntries handles incoming AppendEntries RPC requests on followers.
// This implements receive-heartbeat functionality (of receiving node).
func (n *Node) AppendEntries(args *AppendEntriesRequest, reply *AppendEntriesResponse) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Failed {
		reply.NodeFailed = true
		return nil // Simulate inactivity while failed
	}

	// Heartbeat check: If Entries is empty, it is a heartbeat
	if len(args.Entries) == 0 {
		reply.Term = n.CurrentTerm
		reply.Success = true
		log.Printf("Node %d received heartbeat from Leader %d", n.Id, args.LeaderID)
		n.ResetStopwatchStartTime()
		return nil
	}

	// Case: If follower's term > leader's term, return false
	// if args.Term < n.CurrentTerm {
	// 	log.Printf("TODO: LEADER TRANSITION TO CANDIDATE. Node %d's term: %d > Leader node %d's term: %d ", n.Id, n.CurrentTerm, args.LeaderID, args.Term)
	// 	reply.Term = n.CurrentTerm
	// 	reply.Success = false
	// 	return nil
	// }

	// Consistency check
	// 1. Check if node has a log entry at index = prevLogIndex
	if args.PrevLogIndex >= 0 && args.PrevLogIndex < len(n.Log) { //0, 1
		// 2. Check if nodes entry at prevLogIndex has the same term as PrevLogTerm
		if n.Log[args.PrevLogIndex].Term == args.PrevLogTerm {
			log.Printf("Node %d passed consistency check.", n.Id)
			log.Printf("PrevLogindex: %d. Length of Entries", args.PrevLogIndex, len(args.Entries))
		} else {
			log.Printf("Node %d failed second consistency check of the term comparison.", n.Id)
			if args.PrevLogIndex <= len(n.Log) {
				n.Log = n.Log[:args.PrevLogIndex] // Added
			}
			reply.Term = n.CurrentTerm
			reply.Success = false
			return nil
		}
	} else if args.PrevLogIndex == 0 && len(args.Entries) == 1 {
		log.Printf("Node %d is appending the first entry. PrevLogIndex is: %d.", n.Id, args.PrevLogIndex)

	} else if args.PrevLogIndex == -1 {
		log.Printf("End Reached, Node %d will append all entries.", n.Id)

		for _, entry := range args.Entries {
			n.Log = append(n.Log, entry)
		}

		if args.LeaderCommit > n.CommitIndex {
			n.CommitIndex = min(args.LeaderCommit, len(n.Log)-1)
			//log.Printf("Node %d has committed log index %d", n.Id, n.CommitIndex)
		}

		reply.Term = n.CurrentTerm
		reply.Success = true
		successChan <- true
		return nil

	} else {
		log.Printf("Node %d failed first consistency check of having a prev log entry. PrevLogIndex is: %d. ", n.Id, args.PrevLogIndex)
		if args.PrevLogIndex != -1 && args.PrevLogIndex <= len(n.Log) {
			n.Log = n.Log[:args.PrevLogIndex] // Added
		}
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
		log.Printf("Node %d is appending the first entry. PrevLogIndex is: %d.", n.Id, args.PrevLogIndex)
	}

	// Append new entries to the followers log
	// If it gets here, it means follower passed the consistency check
	if args.PrevLogIndex == 0 && len(args.Entries) == 1 {
		n.Log = append(n.Log, args.Entries[0])
	} else if args.PrevLogIndex == 0 && len(args.Entries) == 2 {
		n.Log = append(n.Log, args.Entries[1])
	} else {
		for _, entry := range args.Entries[1:] {
			n.Log = append(n.Log, entry)
		}
	}
	n.CurrentTerm = args.Entries[len(args.Entries)-1].Term

	if args.LeaderCommit > n.CommitIndex {
		n.CommitIndex = min(args.LeaderCommit, len(n.Log)-1)
		//log.Printf("Node %d has committed log index %d", n.Id, n.CommitIndex)
	}

	reply.Term = n.CurrentTerm
	reply.Success = true
	successChan <- true
	//log.Printf("Node %d successfully appended entries", n.Id)

	return nil
}

// SendAppendEntries sends AppendEntries RPC to a specific follower.
func (leader *Node) SendAppendEntries(peerID int, entries []LogEntry) error {
	leader.mu.Lock()
	log.Printf("Sending AppendEntries to Node %d. Current nextIndex: %v. Current leader's log: %v", peerID, leader.NextIndex, entries)

	var prevLogIndex, prevLogTerm int
	if leader.NextIndex[peerID] == 0 && len(entries) == 1 {
		prevLogIndex = 0
		prevLogTerm = 0
	} else {
		prevLogIndex = leader.NextIndex[peerID] - 1
		if prevLogIndex >= 0 {
			prevLogTerm = leader.Log[prevLogIndex].Term
		}
	}

	if len(entries) != 0 { // If it is not a heartbeat
		// entries = leader.Log[leader.NextIndex[peerID]:] // Changed
		if prevLogIndex == -1 {
			entries = leader.Log
		} else {
			entries = leader.Log[prevLogIndex:]
		}

		// log.Printf("%d", leader.NextIndex[peerID])
		// log.Printf("%d", len(entries))
	}

	args := &AppendEntriesRequest{
		Term:         leader.CurrentTerm,
		LeaderID:     leader.Id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: leader.CommitIndex,
	}
	leader.mu.Unlock()

	var reply AppendEntriesResponse
	err := leader.CallAppendEntriesRPC(peerID, args, &reply)
	if err != nil {
		log.Printf("Leader Node %d failed to send AppendEntries to Node %d: %v\n", leader.Id, peerID, err)
		return nil
	}

	// If Node fails, pause
	if reply.NodeFailed {
		time.Sleep(5 * time.Second)
		// Retry AppendEntries
		//leader.SendAppendEntries(peerID, leader.Log)
		return nil
	}

	// Handle Success Update NextIndex and MatchIndex
	if reply.Success {
		if len(args.Entries) == 0 {
			//log.Printf("Leader Node has sent heartbeat to Node %d", peerID)
		} else {
			//log.Printf("Leader Node %d successfully replicated entries to Node %d", leader.Id, peerID)
			leader.mu.Lock()
			leader.NextIndex[peerID] = len(leader.Log)
			leader.mu.Unlock()
		}
	} else { //Handle Failure
		log.Printf("reply.Success = fail")
		log.Printf("Leader Node %d: AppendEntries to Node %d failed due to inconsistency. Retrying with decremented NextIndex.\n", leader.Id, peerID)
		leader.mu.Lock()
		if leader.NextIndex[peerID] > 0 {
			leader.NextIndex[peerID]--
		}
		leader.mu.Unlock()

		// Retry AppendEntries
		leader.SendAppendEntries(peerID, leader.Log)
	}
	return nil
}

// CallAppendEntriesRPC performs the actual RPC call to AppendEntries on the follower.
func (n *Node) CallAppendEntriesRPC(peerID int, args *AppendEntriesRequest, reply *AppendEntriesResponse) error {
	address := fmt.Sprintf("app%d:8080", peerID) // Assumes Docker container names as app1, app2, etc.
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		return err
	}
	defer client.Close()

	// Call the RPC method
	return client.Call("SingleNode.AppendEntries", args, reply)
}

// SendHeartbeats sends an empty AppendEntries RPC to all followers as a heartbeat.
func (n *Node) SendHeartbeats() {
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue // Skip self
		}
		// Send an empty AppendEntries request (heartbeat) to the follower
		go n.SendAppendEntries(peerID, []LogEntry{}) // No entries means this is a heartbeat
	}
}

// HandleClientCommand processes a client command received by the leader.
func (n *Node) HandleClientCommand(command *string, reply *bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Append the command to the leader's log
	entry := LogEntry{
		Command: *command, // Dereference the pointer to get the string
		Term:    n.CurrentTerm,
	}
	n.Log = append(n.Log, entry)
	log.Printf("Leader Node %d appended command: %s", n.Id, *command)

	// Goroutine to monitor successes and trigger a function on threshold
	go func() {
		successCount := 0
		for range successChan {
			successCount++
			if successCount >= len(n.Peers)/2+1 {
				// Commit entries number of replies is more than half
				n.CommitEntries()
				successCount = 0
			}
		}
	}()

	// Send AppendEntries to all followers with the new entry
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue // Skip sending to self
		}
		go n.SendAppendEntries(peerID, n.Log)
	}

	// Indicate success to the client
	*reply = true
	return nil
}

func (n *Node) CommitEntries() {
	n.CommitIndex += 1
	//log.Printf("Commitindex is now %d", n.CommitIndex)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
