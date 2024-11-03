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
		Id:               id,
		State:            Follower,
		CurrentTerm:      0,
		VotedFor:         -1,
		ElectionTimeout:  time.Duration(rand.Intn(150)+150) * time.Millisecond,
		HeartbeatTimeout: 100 * time.Millisecond,
		Peers:            peers,
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
		// Start sending heartbeats to maintain leadership
		//n.startHeartbeats()
	}
}

// Transit from Follower to Candidate

// Transit from Leader to Follower
