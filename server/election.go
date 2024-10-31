// Leader election logic
package server

// startElection -- node has started an election
// Become candidate, start randomised election timer,  send requestVote RPCs, vote for itself, count votes
// Return an int to let main function know if need to become leader, return to follower, or restart election -- intention is to end the function instead of have nested startElection functions
func (n *Node) startElection() {
	n.state = Candidate
	n.currentTerm += 1

	// Question: can you start an election for a term that someone else is already electing for?
	// If yes, maybe have a mechanism such that you do not start if you have already voted for someone with same term
	// Node 1 -> request vote -> node 2
	// Node 2 -> vote for node 1 
	// Node 2 realise no heartbeat
	// Node 2 --> start election
		// node 2 doesnt bother check and start election => cannot vote for itself
		// node 2 needs to check if it has an availale vote before it can vote for itself
	n.votedFor = n.id

	// Start randomised election timer (150-300ms)

	// Send requestVote RPCs to everyone (concurrently)

	electionResult := n.countVotesReceived()
	if electionResult == 1 {
		n.becomeLeader()
	} else if electionResult == 0 {
		n.becomeFollower(appendEntriesRPC.currentTerm)
	} else if electionResult == -1 {
		go n.startElection()
	}
}

// countVotesReceived
func (n *Node) countVotesReceived() {
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
}
