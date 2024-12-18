package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type FileRequest struct {
	Filename string
}

type FileResponse struct {
	Content []byte
}

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
	startupDelay := 6 * time.Second
	log.Printf("Waiting %v for nodes to start up...", startupDelay)
	time.Sleep(startupDelay)

	// List of nodes in the cluster
	nodes := []string{"app0:8080", "app1:8080", "app2:8080", "app3:8080", "app4:8080"}
	var leader string
	var client *rpc.Client
	var err error

	user := os.Getenv("NAME")

	// Error testing for file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	log.Printf("Current working directory: %s", dir)

	// Initial leader discovery
	leader, client, err = findLeader(nodes) // Leader is a string. Address
	if err != nil {
		log.Fatalf("Could not find the leader after checking all nodes: %v", err)
	}
	defer client.Close()

	// Found leader

	// should then demand the leader to update the client for updated logs
	log.Printf("Loading chat...")
	SendChatLogsRPC(leader)

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
		var reply bool
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
			GetChatLogsRPC(client, leader, commandType)
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

			log := fmt.Sprintf("%s: %s", user, commandMsg)

			// Call SingleNode.HandleClientWrite for Write commands
			err = client.Call("SingleNode.HandleClientWrite", log, &reply)

			time.Sleep(1 * time.Second)
			GetChatLogsRPC(client, leader, commandType)
		}

		// Handle RPC errors
		if err != nil {
			log.Printf("Failed to send command to leader (%s): %v", leader, err)
			client.Close()
			client = nil // Reset the client to trigger reconnection
			continue
		}
	}
}

func GetChatLogsRPC(client *rpc.Client, leader string, commandType string) {
	// Call SingleNode.HandleClientRead for Read commands
	var response string
	err := client.Call("SingleNode.HandleClientRead", &commandType, &response)
	if err != nil {
		log.Printf("Failed to send command to leader (%s): %v", leader, err)
		return
	}

	var logs []struct {
		Term    int    `json:"term"`
		Command string `json:"command"`
	}

	// Parse the JSON response
	err = json.Unmarshal([]byte(response), &logs)
	if err != nil {
		log.Printf("Failed to parse logs: %v", err)
		return
	}

	log.Printf("Logs from leader: %v", logs)

	// Append logs to a file
	file, err := os.Create("logs/logs.txt")
	if err != nil {
		log.Printf("Failed to open logs.txt: %v", err)
		return
	}
	defer file.Close()

	// Write each log entry to the file
	for _, logEntry := range logs {
		_, err := file.WriteString(fmt.Sprintf("Term: %d, %s\n", logEntry.Term, logEntry.Command))
		if err != nil {
			log.Printf("Failed to write to logs.txt: %v", err)
			break
		}
	}
	log.Println("Logs successfully appended to logs.txt")
}

func SendChatLogsRPC(leaderID string) string { // Assume we know the ID of the leader
	leader, err := rpc.Dial("tcp", leaderID)
	if err != nil {
		return fmt.Sprintf("Error connecting to leader: %v", err)
	}
	defer leader.Close()

	fileRequest := FileRequest{Filename: "example.txt"}

	var fileResponse FileResponse

	err = leader.Call("SingleNode.GetFile", &fileRequest, &fileResponse)
	if err != nil {
		fmt.Println("Error calling method:", err)
		return fmt.Sprintf("Error calling method: %v", err)
	}

	fmt.Println("File content:")
	fmt.Println(string(fileResponse.Content))
	return string(fileResponse.Content)
}
