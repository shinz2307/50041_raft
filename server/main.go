package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	lastCommand string // To track the previous command (e.g., W)
	mu sync.Mutex
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}

		// Trim and process the body
		trimmedBody := strings.TrimSpace(string(body))
		log.Printf("Received POST request with body: %s", trimmedBody)

		mu.Lock()
		defer mu.Unlock()

		if trimmedBody == "R" {
			// Handle the read operation
			content, err := os.ReadFile("example.txt")
			if err != nil {
				log.Printf("Error reading example.txt: %v", err)
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, "File content:\n%s", content)
		} else if trimmedBody == "W" {
			// Handle the write operation (track state for the next request)
			lastCommand = "W"
			log.Printf("Tracking W command. Waiting for the next request.")
			fmt.Fprintf(w, "Server is ready for the next input to uppercase.")
		} else if lastCommand == "W" {
			// If the last command was W, append whatever is being written to example.txt
			uppercased := strings.ToUpper(trimmedBody)
		
			// Open the file in append mode (create it if it doesn't exist)
			file, err := os.OpenFile("example.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("Error opening example.txt: %v", err)
				http.Error(w, "Error opening file", http.StatusInternalServerError)
				return		
			}
			defer file.Close()
		
			_, err = file.WriteString(uppercased + "\n")  // Append with a newline
			if err != nil {
				log.Printf("Error writing to example.txt: %v", err)
				http.Error(w, "Error writing to file", http.StatusInternalServerError)
				return
			}
		
			lastCommand = "" // Reset the state
			log.Printf("Uppercased input: %s", uppercased)
			fmt.Fprintf(w, "Uppercased message appended: %s", uppercased)
		
		} else {
			// Handle unexpected input
			http.Error(w, "Unexpected input. Send R or W first.", http.StatusBadRequest)
		}
	} else {
		// Respond to non-POST methods if needed
		fmt.Fprintf(w, "Server is running!")
	}
}



func main() {
	// Register the handler for the "/" route
	http.HandleFunc("/", handler)

	// Log that the server is starting
	log.Println("Server is running on http://0.0.0.0:8080")

	// Start the server and listen on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
