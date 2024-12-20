// server/heartbeatRPC.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"os"
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
	//Case: If follower's term > leader's term, return false
	if  n.CurrentTerm > args.Term {
		log.Printf("Node %d's term: %d > Leader node %d's term: %d ", n.Id, n.CurrentTerm, args.LeaderID, args.Term)
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}
	// Heartbeat check: If Entries is empty, it is a heartbeat
	if len(args.Entries) == 0 {
		n.CurrentTerm = args.Term //test setting followers term same as leader
		reply.Term = n.CurrentTerm
		reply.Success = true
		log.Printf("Node %d at term %d received heartbeat from Leader %d at term%d", n.Id, n.CurrentTerm,args.LeaderID,args.Term)
		n.ResetStopwatchStartTime()
		return nil
	}

	

	// Consistency check
	// 1. Check if node has a log entry at index = prevLogIndex
	if args.PrevLogIndex >= 0 && args.PrevLogIndex < len(n.Log) { //0, 1
		// 2. Check if nodes entry at prevLogIndex has the same term as PrevLogTerm
		if n.Log[args.PrevLogIndex].Term == args.PrevLogTerm {
			log.Printf("Node %d passed consistency check.", n.Id)
			log.Printf("PrevLogindex: %d. Length of Entries: %d", args.PrevLogIndex, len(args.Entries))
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
	//log.Printf("Sending AppendEntries to Node %d. Current nextIndex: %v. Current leader's log: %v", peerID, leader.NextIndex, entries)

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
		//log.Printf("Leader Node %d failed to send AppendEntries to Node %d: %v\n", leader.Id, peerID, err)
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
			leader.MatchIndex[peerID] = len(leader.Log) - 1 // Highest replicated entry
			leader.mu.Unlock()

			// Update commitIndex after successful replication
			leader.UpdateCommitIndex()
		}
	}
		
	if !reply.Success { //Handle Failure
		log.Printf("reply.Success = fail")

		// Failed because of term 
		if reply.Term > leader.CurrentTerm{
			log.Printf("Leader Node %d's term is outdated", leader.Id)
			leader.BecomeFollower(reply.Term)
			return nil
		}

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
func (leader *Node) UpdateCommitIndex() {
	leader.mu.Lock()
	defer leader.mu.Unlock()

	// log.Printf("leader.CommitIndex: %d\n", leader.CommitIndex)
	// log.Printf("leader.log: %d\n", len(leader.Log))
	for N := leader.CommitIndex; N < len(leader.Log); N++ {
		matchCount := 1 // Leader counts itself
		for _, peerID := range leader.Peers {
			if peerID == leader.Id {
				continue
			}
			if leader.MatchIndex[peerID] >= N {
				matchCount++
			}
		}
		// log.Printf("matchCount: %d\n", matchCount)

		// Check majority and current term
		if matchCount > len(leader.Peers)/2 && leader.Log[N].Term == leader.CurrentTerm {
			leader.CommitIndex = N
			log.Printf("Leader Node %d updated commitIndex to %d", leader.Id, leader.CommitIndex)
			leader.ApplyCommittedEntries()
			// leader.SendChatLogsRPC(leader.Id)
		}
	}
}

func (leader *Node) ApplyCommittedEntries() {
	for leader.LastApplied < leader.CommitIndex {
		leader.LastApplied++
		//entry := leader.Log[leader.LastApplied]
		//log.Printf("Leader Node %d applied log entry at index %d: %s", leader.Id, leader.LastApplied, entry.Command)
		// Optionally write to persistent storage here

		file, err := os.OpenFile("example.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		message := leader.Log[leader.CommitIndex].Command
		text := fmt.Sprintf("%v", message)

		if _, err := file.Write([]byte(text)); err != nil {
			log.Fatal(err)
		}
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
		file.Close()

	}
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
	if n.Failed {
		log.Printf("Node %d: Simulating failure. No heartbeats sent.\n", n.Id)
		return
	}
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue // Skip self
		}
		// Send an empty AppendEntries request (heartbeat) to the follower
		go n.SendAppendEntries(peerID, []LogEntry{}) // No entries means this is a heartbeat
	}
}

// HandleClientCommand processes a client command received by the leader.
func (n *Node) HandleClientWrite(command *string, reply *bool) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	log.Printf("HandleClientWrite\n")
	// Append the command to the leader's log
	entry := LogEntry{
		Command: *command, // Dereference the pointer to get the string
		Term:    n.CurrentTerm,
	}
	n.Log = append(n.Log, entry)
	log.Printf("Leader Node %d appended command: %s", n.Id, *command)

	// // Goroutine to monitor successes and trigger a function on threshold
	// go func() {
	// 	if n.Success >= len(n.Peers)/2+1 {
	// 		// Commit entries number of replies is more than half
	// 		n.CommitEntries()
	// 		n.Success = 0
	// 	}
	// }()

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

// HandleClientRead processes a client read request.
// For now, it simply logs the request and returns success.
func (n *Node) HandleClientRead(command *string, reply *string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// log.Printf("HandleClientRead: Received command: %s", *command)

	// Convert logs to JSON
	logsJSON, err := json.Marshal(n.Log)
	if err != nil {
		return err
	}

	// For now, just log the read request and return success.
	// *reply = n.Log
	*reply = string(logsJSON) // Send logs as a JSON string
	return nil
}

// func (n *Node) CommitEntries() {
// 	n.mu.Lock()
// 	defer n.mu.Unlock()

// 	// Ensure we commit only up to the available log entries
// 	if n.CommitIndex < len(n.Log) {
// 		// Get the log entry to commit
// 		committedEntry := n.Log[n.CommitIndex]

// 		// Open the log.txt file for appending
// 		file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err != nil {
// 			log.Printf("Node %d failed to open log.txt: %v", n.Id, err)
// 			return
// 		}
// 		defer file.Close()

// 		// Write the log entry to the file
// 		_, err = file.WriteString(fmt.Sprintf("Term: %d, Command: %s\n", committedEntry.Term, committedEntry.Command))
// 		if err != nil {
// 			log.Printf("Node %d failed to write to log.txt: %v", n.Id, err)
// 			return
// 		}

// 		log.Printf("Node %d committed log entry: %v", n.Id, committedEntry)

// 		// Increment the commit index
// 		n.CommitIndex++
// 	} else {
// 		log.Printf("Node %d has no new entries to commit", n.Id)
// 	}
// }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
