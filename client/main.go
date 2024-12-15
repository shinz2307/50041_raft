package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	for {
		// Read input from the user
		fmt.Print("Enter text to send to the server (or type 'exit' to quit): ")
		var input string
		fmt.Scanln(&input)

		if input == "exit" {
			fmt.Println("Exiting...")
			break
		}

		// Create a POST request with the input text as the body
		resp, err := http.Post("http://app2:8080", "text/plain", bytes.NewBuffer([]byte(input)))
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
		}
		defer resp.Body.Close()

		// Read the response from the server
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Error reading response: %v", err)
		}

		// Print the response
		fmt.Printf("Response from server: %s\n", body)
	}
}
