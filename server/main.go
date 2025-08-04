package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/cloudwego/eino/compose"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan []byte)            // Broadcast channel
var mutex = &sync.Mutex{}                    // Protect clients map

var agent compose.Runnable[string, string] // The chat agent

var registerTypesOnce sync.Once

func init() {
	registerSerializableTypes()
}

func registerSerializableTypes() {
	registerTypesOnce.Do(func() {
		if err := compose.RegisterSerializableType[state]("my state"); err != nil {
			log.Fatal(err)
		}
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	mutex.Lock()
	clients[conn] = true
	mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
		broadcast <- message
	}
}

// formatForLLMOutputWindow replaces newlines with <br> and wraps the text in a <pre> tag for LLM output
func formatForLLMOutputWindow(s string) string {
	return "<pre>" + strings.ReplaceAll(s, "\n", "\n") + "</pre>"
}

// Handles a single user message using the agent and returns the response string
func handleUserMessage(ctx context.Context, userInput string) string {
	result, err := agent.Invoke(ctx, userInput,
		compose.WithCheckPointID("1"),
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, s any) error {
			s.(*state).UserInput = userInput
			return nil
		}),
		compose.WithRuntimeMaxSteps(20),
	)
	info, ok := compose.ExtractInterruptInfo(err)
	if ok {
		s := info.State.(*state)
		response := s.History[len(s.History)-1].Content
		fmt.Printf("Ergebnis bei der Verarbeitung der Benutzereingabe '%s': %+v\n", userInput, err)
		logResponse(userInput, response) // Log the response
		return formatForLLMOutputWindow(response)
	}
	if err != nil {
		response := "[ChatModel error]: " + err.Error()
		fmt.Printf("Fehler bei der Verarbeitung der Benutzereingabe '%s': %+v\n", userInput, err)
		logResponse(userInput, response) // Log the error response
		return formatForLLMOutputWindow(response)
	}
	return formatForLLMOutputWindow(result)
}

func handleMessages() {
	ctx := context.Background()
	for {
		message := <-broadcast
		userInput := string(message)
		response := handleUserMessage(ctx, userInput)

		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(response))
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}

func main() {
	// Load .env file
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Warning: .env file not found or could not be loaded.")
	}

	// Check if OPENAI_API_KEY is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		fmt.Println("Error: OPENAI_API_KEY is not set. Please set it in the .env file.")
		return
	}

	var err error

	agent = createAgent()

	http.HandleFunc("/ws", wsHandler)
	go handleMessages() // optional, falls Broadcast benötigt
	fmt.Println("WebSocket server started on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
