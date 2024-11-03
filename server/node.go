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
	State State
	Id    int

	// Persistent states
	CurrentTerm int
	VotedFor    int
	Log         []LogEntry

	// Volatile states
	CommitIndex int
	LastApplied int
	LeaderID    int

	// Volatile leader states
	NextIndex  map[int]int
	MatchIndex map[int]int

	// RPC handling fields
	HeartbeatTimeout time.Duration
	ElectionTimeout  time.Duration
	Peers            []int
}
