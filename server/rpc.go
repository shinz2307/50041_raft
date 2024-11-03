//go:build exclude
// +build exclude

// RPC handling (RequestVote, AppendEntries)
package server

import "log"

// requestVote RPC request from candidate
type requestVoteRPC struct {
	term         int // Candidate's term
	candidateID  int // Candidate's ID
	lastLogIndex int // Index of the candidate's last log entry
	lastLogTerm  int // Term of the candidate's last log entry
}

type requestVoteReplyRPC struct {
	term        int  // Current term of the responding node, for candidate to update if necessary
	voteGranted bool // True if the candidate received the vote
}

// receiveRequestVote -- process requestVoteRPC and make decision to vote
func (n *Node) receiveRequestVote(rpc requestVoteRPC) {
	log.Printf("Node %d: Received requestVote RPC from node %d", n.id, rpc.candidateID)
	// process and send requestVoteReplyRPC
}

// requestVote -- send out requestVoteRPC
func (n *Node) requestVote() {
	log.Printf("Node %d: Requesting votes for term %d", n.id, n.currentTerm)
	// for each peer, send a requestVoteRPC
}
