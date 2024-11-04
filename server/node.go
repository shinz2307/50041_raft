// Node struct and state (Leader, Follower, Candidate)

package server

import (
	"sync"
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
	State State
	Id    int

	// Persistent states
	CurrentTerm int
	VotedFor    int
	VoteCount int
	Log         []LogEntry

	// Volatile states
	CommitIndex int
	LastApplied int
	LeaderID    int

	// Volatile leader states
	NextIndex  map[int]int
	MatchIndex map[int]int

	// RPC handling fields
	HeartbeatInterval time.Duration
	ElectionTimeout   time.Duration
	Peers             []int

	QuitChannel      <-chan struct{} // Channel to signal node to stop
	resetTimeoutChan chan struct{}   // For followers to reset election timeout
	mu               sync.Mutex
}

