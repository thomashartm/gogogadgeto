package main

import "fmt"

func logResponse(userInput, response string) {
	fmt.Printf("User Input: %s\nLLM Response: %s\n", userInput, response)
}

func logMessage(message string) {
	fmt.Printf("LOG: %s\n", message)
}
