// Persistent state and state transitions
package server

import (
	"fmt"
	"log"
)

// NewNode -- initialise a new node in the Follower state
func NewNode(id int, peers []int) *Node {
	return &Node{
		Id:                id,
		State:             Follower,
		CurrentTerm:       0,
		VotedFor:          -1,
		Peers:             peers,
		resetTimeoutChan:  make(chan struct{}, 1), // Buffered channel
	}
}

// becomeFollower -- Transit from Candidate to Follower
func (n *Node) BecomeFollower(term int) {
	log.Printf("Node %d: Becoming Follower for term %d", n.Id, term)
	n.SetState(Follower)
	n.SetCurrentTerm(term)
	n.SetVotedFor(-1)

	n.RunAsFollower()
}

// becomeLeader -- Transit from Candidate to Leader
func (n *Node) BecomeLeader() {
	if n.State == Candidate {
		log.Printf("Node %d: Transitioning to Leader for term %d", n.Id, n.CurrentTerm)
		n.IncrementCurrentTerm()
		n.SetState(Leader)
		n.SetLeaderID(n.Id)
		n.SetVoteCount(0)
		n.RunAsLeader()

	} else if n.State == Follower { // Follower cannot become leader
        panic(fmt.Sprintf("Node %d cannot become Leader while in Follower state", n.Id))
    }
}

// Follower to Candidate
func (n *Node) BecomeCandidate() {
	//! : StartElection logic is moved to here
	if n.State == Follower || n.State == Candidate {
		log.Printf("Node %d: Transitioning to Candidate for term %d", n.Id, n.CurrentTerm)
		n.SetState(Candidate)
		n.IncrementCurrentTerm()
		n.SetVotedFor(n.Id)
		n.SetVoteCount(1)

		n.SendRequestVoteRPCs()

		n.RunAsCandidate()
	} else if n.State == Leader { // Leader cannot become candidate
        panic(fmt.Sprintf("Node %d cannot become Candidate while in Leader state", n.Id))
    }
}

// Transit from Leader to Follower
