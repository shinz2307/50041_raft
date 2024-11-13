// server/heartbeatRPC.go

package server

import (
	"fmt"
	"log"
	"net/rpc"
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
	Term    int  // Current term for leader to update itself
	Success bool // True if follower contained entry matching PrevLogIndex and PrevLogTerm
}

// AppendEntries handles incoming AppendEntries RPC requests on followers.
func (n *Node) AppendEntries(args *AppendEntriesRequest, reply *AppendEntriesResponse) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Heartbeat check: If Entries is empty, it’s a heartbeat
	if len(args.Entries) == 0 {
		reply.Term = n.CurrentTerm
		reply.Success = true
		//log.Printf("Node %d received heartbeat from Leader %d", n.Id, args.LeaderID)
		return nil
	}

	// Case: If follower's term > leader's term, return false
	if args.Term < n.CurrentTerm {
		log.Printf("Node %d's term: %d > Leader node %d's term: %d ", n.Id, n.CurrentTerm, args.LeaderID, args.Term)
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}

	
	
	// Consistency check 
	// 1. Check if node has a log entry at index = prevLogIndex
	if args.PrevLogIndex >= 0 && args.PrevLogIndex < len(n.Log){
		// 2. Check if nodes entry at prevLogIndex has the same term as PrevLogTerm
		if n.Log[args.PrevLogIndex].Term == args.PrevLogTerm{
			log.Printf("Node %d passed consistency check.", n.Id, )
		}else{
			log.Printf("Node %d failed second consistency check of the term comparison. It has log entries %d", n.Id, n.Log)
			reply.Term = n.CurrentTerm
			reply.Success = false
			return nil
		}		
	}else if args.PrevLogIndex == -1{
		log.Printf("Node %d is appending the first entry. PrevLogIndex is: %d. It has log entries %d", n.Id, args.PrevLogIndex, n.Log)
		
		}else{
		log.Printf("Node %d failed first consistency check of having a prev log entry. PrevLogIndex is: %d. It has log entries %d", n.Id, args.PrevLogIndex, n.Log)
		reply.Term = n.CurrentTerm
		reply.Success = false
		return nil
	}
	
	// TODO: Append new entries to the followers log 
	// If it gets here, it means follower passed the consistency check
	n.Log = append(n.Log,args.Entries...)

	if args.LeaderCommit > n.CommitIndex {
		n.CommitIndex = min(args.LeaderCommit, len(n.Log)-1)
		//log.Printf("Node %d has committed log index %d", n.Id, n.CommitIndex)
	}

	

	reply.Term = n.CurrentTerm
	reply.Success = true
	successChan <- true
	log.Printf("Node %d successfully appended entries", n.Id)

	
	return nil
}

// SendAppendEntries sends AppendEntries RPC to a specific follower.
func (leader *Node) SendAppendEntries(peerID int, entries []LogEntry) {
	leader.mu.Lock()

	var prevLogIndex, prevLogTerm int
	if len(leader.Log) > 0 {
		prevLogIndex = leader.NextIndex[peerID] - 1
		if prevLogIndex >= 0 {
			prevLogTerm = leader.Log[prevLogIndex].Term
		}
	}
	
	if len(entries) !=0{ // If it is not a heartbeat
		entries = leader.Log[leader.NextIndex[peerID]:]
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
		return
	}

	// Handle Success
	//Update NextIndex and MatchIndex 
	//leader.NextIndex[n.Id] = len(n.Log)
	if reply.Success {
		if len(args.Entries) == 0 {
			log.Printf("Leader Node has sent heartbeat to Node %d", peerID)
		} else {
			log.Printf("Leader Node %d successfully replicated entries to Node %d", leader.Id, peerID)
			log.Printf("Peer id: %d. Leader's nextIndex: %v.Length of leader log: %d", peerID, leader.NextIndex,len(leader.Log))
			leader.NextIndex[peerID]=len(leader.Log)
		}
	} else{ //Reply.Success = fail
		log.Printf("reply.Success = fail")
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

	// Goroutine to monitor successes and trigger a function on threshold
	go func() {
		successCount := 0
		for range successChan {
			successCount++
			log.Printf("Current success counter is %d", successCount)
			if successCount >= len(n.Peers)/2+1 {
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
}

func (n *Node) CommitEntries() {
	n.CommitIndex += 1
	log.Printf("Commitindex is now %d", n.CommitIndex)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
