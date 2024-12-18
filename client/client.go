package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"time"
)

// findLeader tries to discover the leader by iterating through all nodes.
func findLeader(nodes []string) (string, *rpc.Client, error) {
	for _, address := range nodes {
		client, err := rpc.Dial("tcp", address)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", address, err)
			continue
		}

		// Check if this node is the leader
		var isLeader bool
		err = client.Call("SingleNode.IsLeader", &struct{}{}, &isLeader)
		if err != nil {
			log.Printf("Failed to call IsLeader on %s: %v", address, err)
			client.Close()
			continue
		}

		if isLeader {
			log.Printf("Found leader: %s", address)
			return address, client, nil
		}
		client.Close() // Close connection if not the leader
	}
	return "", nil, fmt.Errorf("could not find the leader")
}

func main() {
	// Add a startup delay to allow nodes to initialize
	startupDelay := 10 * time.Second
	log.Printf("Waiting %v for nodes to start up...", startupDelay)
	time.Sleep(startupDelay)

	// List of nodes in the cluster
	nodes := []string{"app0:8080", "app1:8080", "app2:8080", "app3:8080", "app4:8080"}
	var leader string
	var client *rpc.Client
	var err error

	// Initial leader discovery
	leader, client, err = findLeader(nodes)
	if err != nil {
		log.Fatalf("Could not find the leader after checking all nodes: %v", err)
	}
	defer client.Close()

	// Interactive loop for user input
	for {
		// Read input from the user
		fmt.Print("Enter command to send to the leader ('R' or 'W' followed by text, or 'exit' to quit): ")
		var commandType string
		fmt.Scanln(&commandType)

		// Exit condition
		if strings.ToLower(commandType) == "exit" {
			fmt.Println("Exiting...")
			break
		}

		// Check if the commandType starts with 'R' (Read) or 'W' (Write)
		if len(commandType) == 0 || (commandType[0] != 'R' && commandType[0] != 'W') {
			fmt.Println("Invalid input. Please start your command with 'R' or 'W'.")
			continue
		}

		// Handle Read ('R') or Write ('W') command
		var reply string
        if client == nil {
            log.Println("Client connection is nil. Reconnecting to leader...")
            leader, client, err = findLeader(nodes)
            if err != nil {
                log.Printf("Failed to re-discover the leader: %v", err)
                time.Sleep(2 * time.Second) // Wait before retrying
                continue
            }
        }

        if commandType[0] == 'R' {
            // Call SingleNode.HandleClientRead for Read commands
            err = client.Call("SingleNode.HandleClientRead", &commandType, &reply)
        } else if commandType[0] == 'W' {
            // Read write data from the user
            fmt.Print("Enter your message, or 'exit' to quit: ")
            var commandMsg string
            fmt.Scanln(&commandMsg)

            // Exit condition
            if strings.ToLower(commandMsg) == "exit" {
                fmt.Println("Exiting...")
                break
            }

            // Call SingleNode.HandleClientWrite for Write commands
            err = client.Call("SingleNode.HandleClientWrite", &commandMsg, &reply)
        }

        // Handle RPC errors
        if err != nil {
            log.Printf("Failed to send command to leader (%s): %v", leader, err)
            client.Close()
            client = nil // Reset the client to trigger reconnection
            continue
        }

        // Print the response from the leader
        fmt.Printf("Response from leader: %s\n", reply)
	}
}