package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"gogogajeto/agent/chatmodel"
	"gogogajeto/agent/tools"
	"gogogajeto/util"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
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

type AgentResult struct {
	Info      string
	LLMOutput string
}

// Reasoning struct for structured reasoning output
type Reasoning struct {
	BeforeNode      []string          `json:"beforeNode"`
	AfterNode       []string          `json:"afterNode"`
	RerunNode       []string          `json:"rerunNode"`
	RerunNodesExtra map[string]string `json:"rerunNodesExtra"`
	SubGraphs       map[string]string `json:"subGraphs"`
}

// HistoryItem represents a single message in the conversation history
type HistoryItem struct {
	OrderID    int            `json:"orderId"`
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []ToolCallInfo `json:"toolCalls,omitempty"`
	ToolCallID string         `json:"toolCallId,omitempty"`
	Name       string         `json:"name,omitempty"`
}

// ToolCallInfo represents a tool call
type ToolCallInfo struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionInfo `json:"function"`
}

// FunctionInfo represents function call details
type FunctionInfo struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func init() {
	registerSerializableTypes()
}

func registerSerializableTypes() {
	registerTypesOnce.Do(func() {
		if err := compose.RegisterSerializableType[chatmodel.State]("my state"); err != nil {
			log.Fatal(err)
		}
	})
}

// mapInterruptInfoToReasoning converts InterruptInfo to our Reasoning struct
func mapInterruptInfoToReasoning(info *compose.InterruptInfo) Reasoning {
	reasoning := Reasoning{
		BeforeNode:      []string{},
		AfterNode:       []string{},
		RerunNode:       []string{},
		RerunNodesExtra: make(map[string]string),
		SubGraphs:       make(map[string]string),
	}

	if info == nil {
		return reasoning
	}

	// Map BeforeNodes directly
	if info.BeforeNodes != nil {
		reasoning.BeforeNode = info.BeforeNodes
	}

	// Map AfterNodes directly
	if info.AfterNodes != nil {
		reasoning.AfterNode = info.AfterNodes
	}

	// Map RerunNodes directly
	if info.RerunNodes != nil {
		reasoning.RerunNode = info.RerunNodes
	}

	// Convert RerunNodesExtra from map[string]any to map[string]string
	if info.RerunNodesExtra != nil {
		for key, value := range info.RerunNodesExtra {
			if value != nil {
				reasoning.RerunNodesExtra[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	// Convert SubGraphs from map[string]*InterruptInfo to map[string]string
	if info.SubGraphs != nil {
		for key, subGraph := range info.SubGraphs {
			if subGraph != nil {
				// Create a summary of the subgraph without including the full state
				summary := fmt.Sprintf("SubGraph{BeforeNodes:%d,AfterNodes:%d,RerunNodes:%d}",
					len(subGraph.BeforeNodes),
					len(subGraph.AfterNodes),
					len(subGraph.RerunNodes))
				reasoning.SubGraphs[key] = summary
			}
		}
	}

	return reasoning
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

func formatAsJsonForLLMOutputWindow(response string, info *compose.InterruptInfo, history []*schema.Message) string {
	type Output struct {
		Response  string        `json:"response"`
		Reasoning Reasoning     `json:"reasoning"`
		History   []HistoryItem `json:"history"`
	}

	reasoning := mapInterruptInfoToReasoning(info)

	// Convert history to structured format
	historyItems := make([]HistoryItem, len(history))
	for i, msg := range history {
		historyItem := HistoryItem{
			OrderID: i,
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		// Add tool calls if present
		if len(msg.ToolCalls) > 0 {
			historyItem.ToolCalls = make([]ToolCallInfo, len(msg.ToolCalls))
			for j, toolCall := range msg.ToolCalls {
				historyItem.ToolCalls[j] = ToolCallInfo{
					ID:   toolCall.ID,
					Type: toolCall.Type,
					Function: FunctionInfo{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				}
			}
		}

		// Add tool call ID if this is a tool response
		if msg.Role == schema.Tool {
			historyItem.ToolCallID = msg.ToolCallID
			historyItem.Name = msg.Name
		}

		historyItems[i] = historyItem
	}

	out := Output{
		Response:  response,
		Reasoning: reasoning,
		History:   historyItems,
	}

	b, err := json.Marshal(out)
	if err != nil {
		return `{"response":"JSON error","reasoning":{"beforeNode":[],"afterNode":[],"rerunNode":[],"rerunNodesExtra":{},"subGraphs":{}},"history":[]}`
	}
	return string(b)
}

// Handles a single user message using the agent and returns the response string
func handleUserMessage(ctx context.Context, userInput string) string {
	util.LogMessage("=== CONVERSATION START ===")
	util.LogMessage("User input: " + userInput)

	// Generate a unique checkpoint ID for each conversation
	checkpointID := strconv.FormatInt(time.Now().UnixNano(), 10)

	result, err := agent.Invoke(ctx, userInput,
		compose.WithCheckPointID(checkpointID),
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, s any) error {
			s.(*chatmodel.State).UserInput = userInput
			return nil
		}),
		compose.WithRuntimeMaxSteps(20),
	)

	util.LogMessage("Agent.Invoke completed")

	info, ok := compose.ExtractInterruptInfo(err)
	if ok {
		util.LogMessage("=== INTERRUPT INFO EXTRACTED ===")
		s := info.State.(*chatmodel.State)

		util.LogMessage(fmt.Sprintf("Conversation history length: %d", len(s.History)))
		if len(s.History) > 0 {
			lastMessage := s.History[len(s.History)-1]
			util.LogMessage("Last message role: " + string(lastMessage.Role))
			util.LogMessage(fmt.Sprintf("Last message content length: %d", len(lastMessage.Content)))
		}

		response := s.History[len(s.History)-1].Content
		util.LogMessage("=== CONVERSATION SUCCESS ===")
		fmt.Printf("Result after processing user input: '%s': %+v\n", userInput, err)
		//util.LogResponse(userInput, response) // Log the response
		return formatAsJsonForLLMOutputWindow(response, info, s.History)
	}
	if err != nil {
		util.LogMessage("=== CONVERSATION ERROR ===")
		util.LogMessage("Error: " + err.Error())
		response := "[ChatModel error]: " + err.Error()
		fmt.Printf("Error while processing user input: '%s': %+v\n", userInput, err)
		util.LogResponse(userInput, response) // Log the error response
		return formatAsJsonForLLMOutputWindow(response, info, nil)
	}

	util.LogMessage("=== CONVERSATION COMPLETED WITHOUT INTERRUPT ===")
	util.LogMessage("Direct result: " + result)
	return formatAsJsonForLLMOutputWindow(result, info, nil)
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

func createAgent() compose.Runnable[string, string] {
	util.LogMessage("=== AGENT CREATION START ===")
	ctx := context.Background()

	// init Python sandbox and tools
	util.LogMessage("Creating Python sandbox...")
	pythonSb := tools.NewSandbox(ctx)
	//defer pythonSb.Cleanup(ctx)

	util.LogMessage("Creating Python command line tools...")
	pythonTools := tools.NewCommandLineTool(ctx, pythonSb)
	util.LogMessage(fmt.Sprintf("Created %d Python tools", len(pythonTools)))

	// init Kali Linux sandbox and tools
	util.LogMessage("Creating Kali Linux sandbox...")
	kaliSb := tools.NewKaliSandbox(ctx)
	//defer kaliSb.Cleanup(ctx)

	util.LogMessage("Creating Kali information gathering tools...")
	kaliTools := tools.NewKaliCommandLineTool(ctx, kaliSb)
	util.LogMessage(fmt.Sprintf("Created %d Kali tools", len(kaliTools)))

	// Combine all tools
	allTools := append(pythonTools, kaliTools...)
	util.LogMessage(fmt.Sprintf("Total tools available: %d", len(allTools)))

	// init chat model and bind tools
	util.LogMessage("Creating chat model...")
	cm := tools.NewChatModel(ctx)

	util.LogMessage("Binding all tools to chat model...")
	cm = tools.BindTools(ctx, cm, allTools)

	// create agent
	util.LogMessage("Composing agent...")
	agent := chatmodel.ComposeAgent(ctx, cm, allTools)

	util.LogMessage("=== AGENT CREATION COMPLETE ===")
	return agent
}

func main() {
	// Print the ASCII art logo
	fmt.Print(Logo)

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
