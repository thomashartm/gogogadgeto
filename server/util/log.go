package util

import "fmt"

func LogResponse(userInput, response string) {
	fmt.Printf("User Input: %s\nLLM Response: %s\n", userInput, response)
}

func LogMessage(message string) {
	fmt.Printf("LOG: %s\n", message)
}
