// Node struct and state (Leader, Follower, Candidate)

package server

import (
	"time"
)

type State int

const (
	// iota: 0, Follower = 0
	Follower State = iota
	// iota: 1, Candidate = 1
	Candidate
	// iota: 2, Leader = 2
	Leader
)

type LogEntry struct {
	Term    int
	Index   int
	Command interface{}
}

type Node struct {
	state State
	id    int

	// Persistent states
	currentTerm int
	votedFor    int
	log         []LogEntry

	// Volatile states
	commitIndex int
	lastApplied int
	leaderID    int

	// Volatile leader states
	nextIndex  map[int]int
	matchIndex map[int]int

	// RPC handling fields
	heartbeatTimeout time.Duration
	electionTimeout  time.Duration
	peers            []int
}
