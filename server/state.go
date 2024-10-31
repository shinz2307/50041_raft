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
		id:					id,
		state:				Follower,
		currentTerm:		0,
		votedFor: 			-1,
		electionTimeout:	time.Duration(rand.Intn(150)+150) * time.Millisecond,
		heartbeatTimeout:	100 * time.Millisecond,
		peers:				peers,
	}
}


// becomeFollower -- Transit from Candidate to Follower
func (n* Node) becomeFollower(term int) {
	log.Printf("Node %d: Becoming Follower for term %d", n.id, term)
    n.state = Follower
    n.currentTerm = term
    n.votedFor = -1
}

// becomeLeader -- Transit from Candidate to Leader
func (n *Node) becomeLeader() {
    if n.state == Candidate {
        log.Printf("Node %d: Transitioning to Leader for term %d", n.id, n.currentTerm)
        n.state = Leader
        // Start sending heartbeats to maintain leadership
        n.startHeartbeats()
    }
}

// Transit from Follower to Candidate


// Transit from Leader to Follower

