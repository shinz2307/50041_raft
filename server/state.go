// Persistent state and state transitions
package server

import (
	"log"
	"math/rand"
	"time"
)

// NewNode -- initialise a new node in the Follower state
func NewNode(id int, peers []int) *Node {
	return &Node{
		Id:                id,
		State:             Follower,
		CurrentTerm:       0,
		VotedFor:          -1,
		ElectionTimeout:   time.Duration(rand.Intn(150)+150) * time.Millisecond,
		HeartbeatInterval: 100 * time.Millisecond,
		Peers:             peers,
		resetTimeoutChan:  make(chan struct{}, 1), // Buffered channel
	}
}

// becomeFollower -- Transit from Candidate to Follower
func (n *Node) BecomeFollower(term int) {
	log.Printf("Node %d: Becoming Follower for term %d", n.Id, term)
	n.State = Follower
	n.CurrentTerm = term
	n.VotedFor = -1
}

// becomeLeader -- Transit from Candidate to Leader
func (n *Node) BecomeLeader() {
	if n.State == Candidate {
		log.Printf("Node %d: Transitioning to Leader for term %d", n.Id, n.CurrentTerm)
		n.State = Leader
		n.LeaderID = n.Id
		// Start sending heartbeats to maintain leadership
		// TODO: For now; i enabled the following but maybe we need an automatic way to send out heartbeats
		// n.SendHeartbeats()
	}
	n.VoteCount = 0
}

// Follower to Candidate
func (n *Node) BecomeCandidate() {
	if n.State == Follower {
		log.Printf("Node %d: Transitioning to Candidate for term %d", n.Id, n.CurrentTerm)
		n.State = Candidate
	}
}

// Transit from Leader to Follower
