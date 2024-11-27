package server

// Receiving RequestVoteRPC logic is here
import (
	"fmt"
	"log"
	"net/rpc"
)

type RequestVoteArgs struct {
	// Pertaining to the candidate
	Term         int
	CandidateID  int 
	LastLogIndex int // Index of last log entry
	LastLogTerm  int // Term of last log entry

}

type RequestVoteReply struct {
	Term int
	VoteGranted bool
}

// Function will run on candidate's
func (n *Node) SendRequestVoteRPCs() {
	log.Printf("Node %d sends out RequestVoteRPCs for term %d", n.Id, n.CurrentTerm)

	var lastLogIndex int
    var lastLogTerm int

    if len(n.Log) > 0 {
        lastLogIndex = len(n.Log) - 1
        lastLogTerm = n.Log[lastLogIndex].Term
    } else { // No log entries
        lastLogIndex = 0
        lastLogTerm = 0
    }

	args := &RequestVoteArgs{
		Term: n.CurrentTerm,
		CandidateID: n.Id,
		LastLogIndex: lastLogIndex, // TODO: CHECK
		LastLogTerm: lastLogTerm, // TODO: CHECK
	}


	// Send RequestVoteRPC to all peers except this
	for _, peerID := range n.Peers {
		if peerID == n.Id {
			continue
		}

		go func(id int) {
			var reply RequestVoteReply
			err := n.CallRequestVoteRPC(id, args, &reply)
			if err != nil {
				log.Printf("Candidate Node %d failed to call RequestVoteRPC on Node %d: %v\n", n.Id, id, err)
			}

			n.CountVote(&reply) // handle the reply
		}(peerID)

	}
}

// Function will run on the candidate's
func (n *Node) CallRequestVoteRPC(peerID int, args *RequestVoteArgs, reply *RequestVoteReply) error {
	address := fmt.Sprintf("localhost:%d", 8000+peerID) // Construct the target address
	log.Printf("Sending RequestVoteRPC to %v\n", address)
	client, err := rpc.Dial("tcp", address) // Establish an RPC connection to the follower node
	if err != nil {                         // Handle connection error
		return err
	}
	defer client.Close() // Ensure the connection is closed after the RPC call

	// Use the unique service name to call the Heartbeat method on the target node
	serviceName := fmt.Sprintf("Node%d.RequestVote", peerID)
	return client.Call(serviceName, args, reply) // Make the RPC call to the follower's RequestVote method
}


// Function will run on follower or ANOTHER candidate's
func (n *Node) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	// log.Printf("Node %d received RequestVoteRPC from Node %d for term %d\n", n.Id, args.CandidateID, args.Term)

	reply.Term = n.CurrentTerm

	if n.CurrentTerm > args.Term {
		reply.VoteGranted = false
		log.Printf("Node %d denies vote to Node %d: current term %d is > the candidate's term %d\n", n.Id, args.CandidateID, n.CurrentTerm, args.Term)
	} else if (n.VotedFor == -1 || n.VotedFor == args.CandidateID) && n.CommitIndex <= args.LastLogIndex {
		// We take -1 as NULL
		// We also use the last log index as a measure for "updated-ness"
		reply.VoteGranted = true
		log.Printf("Node %d votes true to Node %d for term %d.\n", n.Id, args.CandidateID, args.Term)
		n.VotedFor = args.CandidateID
	} else {
		reply.VoteGranted = false

		// Following is just for LOGGING!:
		log.Printf("Node %d votes false to Node %d for term %d. Reasons:\n", n.Id, args.CandidateID, args.Term)
		if n.CurrentTerm <= args.Term {
			log.Printf("\t- The candidate's term > this node (the candidate's term: %d, current term: %d) BUT\n", args.Term, n.CurrentTerm)
		}
		if n.VotedFor != -1 && n.VotedFor != args.CandidateID {
			log.Printf("\t- Already voted for another candidate (voted for: %d, the candidate: %d)\n", n.VotedFor, args.CandidateID)
		}
		if n.CommitIndex > args.LastLogIndex {
			log.Printf("\t- The candidate's log is outdated (the candidate's last log index: %d, current commit index: %d)\n", args.LastLogIndex, n.CommitIndex)
		}

	}


	// log.Printf("Node %d response to Node %d for term %d: voteGranted = %v\n", n.Id, args.CandidateID, reply.Term, reply.VoteGranted)
	return nil // No error occurred
}
