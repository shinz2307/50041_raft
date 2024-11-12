package server

import (
	"log"
)

// startElection -- node has started an election
// Become candidate, start randomised election timer,  send requestVote RPCs, vote for itself, count votes
// Return an int to let main function know if need to become leader, return to follower, or restart election -- intention is to end the function instead of have nested startElection functions
func (n *Node) StartElection() { // TODO: UNUSED NOW. Replaced by BecomeCandidate() instead!
	log.Printf("Node %d is starting an election for term %d\n", n.Id, n.CurrentTerm)

	n.SetState(Candidate)
	n.IncrementCurrentTerm()
	n.SetVotedFor(n.Id)
	n.SetVoteCount(1)

	n.SendRequestVoteRPCs()
	log.Printf("Node %d starts election timer %d\n", n.Id, n.CurrentTerm)
    go n.BeginStateTimer() // Start timer in background

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

	if n.State == Leader { // This is to ignore any more vote counts. This node is already a leader
		log.Printf("Node %d received a vote but is already a leader!", n.Id)
		return
	}

	// So if the incoming reply is from a more updated node, this node transits to follower
	if reply.Term > n.CurrentTerm {
		log.Printf("Node %d received a reply for RequestVoteRPC: but it has a higher term! This node becomes a follower!", n.Id)
		n.BecomeFollower(reply.Term)
		return
	}


	if reply.VoteGranted {
		log.Printf("Node %d has received a VoteGranted!\n", n.Id)

		n.IncrementVoteCount()
		log.Printf("Node %d current vote count: %d\n", n.Id, n.VoteCount)
		if n.VoteCount > len(n.Peers)/2 {
			log.Printf("Node %d has majority of votes!\n", n.Id)
			n.BecomeLeader() 
		}
	}
}