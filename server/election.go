package server

import (
	"log"
	"math/rand"
	"time"
)

const (
	minElectionTimeout = 1500 // milliseconds. For now i put as 1.5 seconds
	maxElectionTimeout = 3000 // milliseconds
)

// startElection -- node has started an election
// Become candidate, start randomised election timer,  send requestVote RPCs, vote for itself, count votes
// Return an int to let main function know if need to become leader, return to follower, or restart election -- intention is to end the function instead of have nested startElection functions
func (n *Node) StartElection() {

	n.State = Candidate
	n.CurrentTerm++
	n.VotedFor = n.Id
	n.VoteCount = 1

	log.Printf("Node %d is starting an election for term %d\n", n.Id, n.CurrentTerm)

	n.StartElectionTimer() // Will recursively call this function StartElection() upon timeout
	n.SendRequestVoteRPCs()

	// Question: can you start an election for a term that someone else is already electing for?
	// If yes, maybe have a mechanism such that you do not start if you have already voted for someone with same term
	// Node 1 -> request vote -> node 2
	// Node 2 -> vote for node 1
	// Node 2 realise no heartbeat
	// Node 2 --> start election
	// node 2 doesnt bother check and start election => cannot vote for itself
	// node 2 needs to check if it has an availale vote before it can vote for itself

	// Start randomised election timer (150-300ms)

	// Send requestVote RPCs to everyone (concurrently)

}

// CountVote
func (n *Node) CountVote(reply *RequestVoteReply) {
	// If vote comes in:
	// Check if vote granted
	// If yes, add to list of voters
	// If number of votes > n/2
	// Become leader: return 1

	// If receive appendEntriesRPC:
	// If appendEntries.currentTerm >= n.currentTerm:
	// Become follower: return 0

	// If timer runs out:
	// n.startElection(): return -1

	// So if the incoming reply is from a more updated node, this node transits to follower
	if reply.Term > n.CurrentTerm {
		n.BecomeFollower(reply.Term)
		return
	}

	if reply.VoteGranted {
		n.VoteCount++

		if n.VoteCount > len(n.Peers)/2 {
			log.Printf("Node %d has received a majority of votes and is now the leader!\n", n.Id)
			n.BecomeLeader()
			return
		}
	}

}

func (n *Node) StartElectionTimer() {
	timerDuration := rand.Intn(maxElectionTimeout-minElectionTimeout) + minElectionTimeout
	time.AfterFunc(time.Duration(timerDuration)*time.Millisecond, func() {
		if n.State != Leader {
			log.Printf("Node %d election timer expired, starting a new election...\n", n.Id)
			// n.StartElection()
		}
	})
}
