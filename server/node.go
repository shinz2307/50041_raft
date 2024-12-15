// Node struct and state (Leader, Follower, Candidate)

package main

import (
	"fmt"
	"log"
	"math/rand"
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

// This function is to print out the state as its string name instead of its integer representation
func (s State) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"
	}
}

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
	VoteCount   int
	Log         []LogEntry

	// Volatile states
	CommitIndex int
	LastApplied int
	LeaderID    int

	// Volatile leader states
	NextIndex  map[int]int
	MatchIndex map[int]int

	// RPC handling fields
	TimeoutOrHeartbeatInterval time.Duration
	// TLDR: one variable for all timeouts / heartbeat interval:
	// Follower will just use heartbeat timeout. Then this variable becomes that
	// Candidate will just use election timeout. Then this variable becomes that
	// Leader will use a heartbeat interval (to periodically send out heartbeats). Then this variable becomes that

	stopwatchStartTime time.Time

	HeartbeatInterval time.Duration
	ElectionTimeout   time.Duration

	Peers          []int
	CommandChannel chan string //channel for client commands

	QuitChannel      <-chan struct{} // Channel to signal node to stop
	Failed           bool
	resetTimeoutChan chan struct{} // For followers to reset election timeout
	mu               sync.Mutex
}
func (n *Node) IsLeader(_ *struct{}, reply *bool) error { 
	n.mu.Lock()
	defer n.mu.Unlock()

	log.Printf("Node %d state: %s", n.Id, n.State.String())
	*reply = (n.State == Leader)
	return nil
   }

func (n *Node) GetLog() []LogEntry {
	n.mu.Lock()
	defer n.mu.Unlock()
	logCopy := make([]LogEntry, len(n.Log))
	copy(logCopy, n.Log)
	return logCopy
}

func (n *Node) SetState(state State) {
	log.Printf("Node %d sets me.State = %s\n", n.Id, state)
	n.State = state
}

func (n *Node) SetCurrentTerm(term int) {
	log.Printf("Node %d sets me.CurrentTerm = %d\n", n.Id, term)
	n.CurrentTerm = term
}

func (n *Node) SetLeaderID(leaderID int) {
	log.Printf("Node %d sets me.LeaderID = %d\n", n.Id, leaderID)
	n.LeaderID = leaderID
}

func (n *Node) SetVotedFor(votedID int) {
	log.Printf("Node %d sets me.VotedFor = Node %d\n", n.Id, votedID)
	n.VotedFor = votedID
}

func (n *Node) SetVoteCount(voteCount int) {
	log.Printf("Node %d sets me.VoteCount = %d\n", n.Id, voteCount)
	n.VoteCount = voteCount
}

func (n *Node) SetTimeoutOrHeartbeatInterval() {
	// printTimer := func() {
	// 	log.Printf("Node %d timeout / heartbeat interval is set to %s\n", n.Id, n.TimeoutOrHeartbeatInterval)
	// }

	switch n.State {
	case Follower:
		n.TimeoutOrHeartbeatInterval = time.Duration((5 + rand.Float64()*2) * float64(time.Second)) // 5-7 seconds
		//printTimer()
	case Candidate:
		n.TimeoutOrHeartbeatInterval = time.Duration((3 + rand.Float64()*2) * float64(time.Second)) // 3-5 seconds
		//printTimer()
	case Leader:
		n.TimeoutOrHeartbeatInterval = time.Duration((1 + rand.Float64()*1) * float64(time.Second)) // 1-2 seconds for heartbeat
		//printTimer()
	default:
		// If a node is not any of the above states. SHOULD NOT HAPPEN
		panic(fmt.Sprintf("Invalid node state: %s", n.State))
	}
}

func (n *Node) IncrementCurrentTerm() {
	log.Printf("Node %d increments its current term by 1\n", n.Id)
	n.CurrentTerm++
}

func (n *Node) IncrementVoteCount() {
	log.Printf("Node %d increments its vote count by 1\n", n.Id)
	n.VoteCount++

}

func (n *Node) ResetStopwatchStartTime() {

	log.Printf("Node %d resets its timeout.\n", n.Id)
	n.stopwatchStartTime = time.Now() // Reset stopwatch to current time
}

func (n *Node) BeginStateTimer() {
	n.SetTimeoutOrHeartbeatInterval()
	log.Printf("Node %d begins its state(%s) timer with timeout: %s\n", n.Id, n.State, GetFormatDuration(n.TimeoutOrHeartbeatInterval))

	n.ResetStopwatchStartTime() // Start stopwatch when timer begins

	timer := time.NewTimer(n.TimeoutOrHeartbeatInterval)
	go func() {
		for {
			select {
			case <-timer.C: // Block the code until this line; when timer expires
				n.mu.Lock()
				elapsedTime := time.Since(n.stopwatchStartTime)
				n.mu.Unlock()

				if n.State == Follower || n.State == Candidate {
					if elapsedTime >= n.TimeoutOrHeartbeatInterval {
						log.Printf("Node %d's stopwatch exceeded timeout. Becoming Candidate.\n", n.Id)
						n.BecomeCandidate()
						// DO NOT RETURN. Because we want to permanently loop.
					}

				} else if n.State == Leader {
					log.Printf("Node %d is the Leader. Sending heartbeat.\n", n.Id)
					n.SendHeartbeats()
					timer.Reset(n.TimeoutOrHeartbeatInterval)
					// DO NOT RETURN. Because we want to permanently loop.
				}

			case <-n.QuitChannel: // This is for manual stopping. If we want we need to implement a quit signal
				timer.Stop()
				return
			}
		}
	}()
}

func (n *Node) PrintNodeTimer() {
	log.Printf("Node %d current timeout / heartbeat interval is %s\n", n.Id, n.TimeoutOrHeartbeatInterval)
}
