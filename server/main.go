package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"gogogajeto/agent/common"
	manus "gogogajeto/agent/manus"
	"gogogajeto/util"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
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

// Session management
var sessions = make(map[string]*SessionInfo) // Active sessions
var sessionMutex = &sync.RWMutex{}           // Protect sessions map

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

// Session management types
type SessionInfo struct {
	SessionID    string    `json:"sessionId"`
	CreatedAt    time.Time `json:"createdAt"`
	LastAccess   time.Time `json:"lastAccess"`
	MessageCount int       `json:"messageCount"`
}

type SessionRequest struct {
	SessionID string `json:"sessionId,omitempty"`
	Message   string `json:"message"`
}

type SessionResponse struct {
	SessionID string        `json:"sessionId"`
	Response  string        `json:"response"`
	History   []HistoryItem `json:"history,omitempty"`
}

func init() {
	registerSerializableTypes()
}

func registerSerializableTypes() {
	registerTypesOnce.Do(func() {
		if err := compose.RegisterSerializableType[common.State]("my state"); err != nil {
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

// Session management functions
func createSession() *SessionInfo {
	sessionID := uuid.New().String()
	session := &SessionInfo{
		SessionID:    sessionID,
		CreatedAt:    time.Now(),
		LastAccess:   time.Now(),
		MessageCount: 0,
	}

	sessionMutex.Lock()
	sessions[sessionID] = session
	sessionMutex.Unlock()

	util.LogMessage(fmt.Sprintf("Created new session: %s", sessionID))
	return session
}

func getSession(sessionID string) (*SessionInfo, bool) {
	sessionMutex.RLock()
	session, exists := sessions[sessionID]
	sessionMutex.RUnlock()

	if exists {
		// Update last access time
		sessionMutex.Lock()
		session.LastAccess = time.Now()
		sessionMutex.Unlock()
	}

	return session, exists
}

func deleteSession(sessionID string) {
	sessionMutex.Lock()
	delete(sessions, sessionID)
	sessionMutex.Unlock()

	util.LogMessage(fmt.Sprintf("Deleted session: %s", sessionID))
}

// Handles a single user message using the agent with session management
func handleUserMessageWithSession(ctx context.Context, sessionID, userInput string) SessionResponse {
	util.LogMessage("=== CONVERSATION START ===")
	util.LogMessage(fmt.Sprintf("Session ID: %s", sessionID))
	util.LogMessage("User input: " + userInput)

	// Update session info
	session, exists := getSession(sessionID)
	if !exists {
		util.LogMessage("Session not found, creating new one")
		session = createSession()
		sessionID = session.SessionID
	}

	session.MessageCount++

	// Use sessionID as checkpoint ID (this is the key fix!)
	result, err := agent.Invoke(ctx, userInput,
		compose.WithCheckPointID(sessionID), // Use sessionID instead of timestamp
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, s any) error {
			s.(*common.State).UserInput = userInput
			return nil
		}),
		compose.WithRuntimeMaxSteps(20),
	)

	util.LogMessage("Agent.Invoke completed")

	response := SessionResponse{
		SessionID: sessionID,
	}

	info, ok := compose.ExtractInterruptInfo(err)
	if ok {
		util.LogMessage("=== INTERRUPT INFO EXTRACTED ===")
		s := info.State.(*common.State)

		util.LogMessage(fmt.Sprintf("Conversation history length: %d", len(s.History)))
		if len(s.History) > 0 {
			lastMessage := s.History[len(s.History)-1]
			util.LogMessage("Last message role: " + string(lastMessage.Role))
			util.LogMessage(fmt.Sprintf("Last message content length: %d", len(lastMessage.Content)))
		}

		responseText := s.History[len(s.History)-1].Content
		response.Response = formatAsJsonForLLMOutputWindow(responseText, info, s.History)

		// Convert history for response
		response.History = make([]HistoryItem, len(s.History))
		for i, msg := range s.History {
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

			response.History[i] = historyItem
		}

		util.LogMessage("=== CONVERSATION SUCCESS ===")
		return response
	}

	if err != nil {
		util.LogMessage("=== CONVERSATION ERROR ===")
		util.LogMessage("Error: " + err.Error())
		responseText := "[ChatModel error]: " + err.Error()
		response.Response = formatAsJsonForLLMOutputWindow(responseText, info, nil)
		return response
	}

	util.LogMessage("=== CONVERSATION COMPLETED WITHOUT INTERRUPT ===")
	util.LogMessage("Direct result: " + result)
	response.Response = formatAsJsonForLLMOutputWindow(result, info, nil)
	return response
}

// Handles a single user message using the agent and returns the response string (legacy function for backward compatibility)
func handleUserMessage(ctx context.Context, userInput string) string {
	util.LogMessage("=== CONVERSATION START ===")
	util.LogMessage("User input: " + userInput)

	// Generate a unique checkpoint ID for each conversation
	checkpointID := strconv.FormatInt(time.Now().UnixNano(), 10)

	result, err := agent.Invoke(ctx, userInput,
		compose.WithCheckPointID(checkpointID),
		compose.WithStateModifier(func(ctx context.Context, path compose.NodePath, s any) error {
			s.(*common.State).UserInput = userInput
			return nil
		}),
		compose.WithRuntimeMaxSteps(20),
	)

	util.LogMessage("Agent.Invoke completed")

	info, ok := compose.ExtractInterruptInfo(err)
	if ok {
		util.LogMessage("=== INTERRUPT INFO EXTRACTED ===")
		s := info.State.(*common.State)

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
	// Create a default session for legacy WebSocket messages that don't specify a session
	defaultSession := createSession()

	for {
		message := <-broadcast
		userInput := string(message)

		// Use session-based handling even for legacy messages
		sessionResponse := handleUserMessageWithSession(ctx, defaultSession.SessionID, userInput)
		response := sessionResponse.Response

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

// HTTP endpoints for session management
func sessionNewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session := createSession()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func sessionMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	response := handleUserMessageWithSession(ctx, req.SessionID, req.Message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sessionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/session/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[1] != "history" {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	sessionID := parts[0]
	session, exists := getSession(sessionID)
	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func sessionDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract session ID from URL path
	sessionID := strings.TrimPrefix(r.URL.Path, "/api/session/")

	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	deleteSession(sessionID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// Enhanced WebSocket handler that supports session-based messaging
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

		// Try to parse as SessionRequest first
		var sessionReq SessionRequest
		if err := json.Unmarshal(message, &sessionReq); err == nil && sessionReq.Message != "" {
			// Handle as session-based message
			ctx := context.Background()
			response := handleUserMessageWithSession(ctx, sessionReq.SessionID, sessionReq.Message)

			responseBytes, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, responseBytes)
		} else {
			// Fall back to legacy message handling for backward compatibility
			broadcast <- message
		}
	}
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

	agent = manus.CreateAgent()

	// Register HTTP endpoints
	http.HandleFunc("/api/session/new", sessionNewHandler)
	http.HandleFunc("/api/session/message", sessionMessageHandler)
	http.HandleFunc("/api/session/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/history") {
			sessionHistoryHandler(w, r)
		} else if r.Method == "DELETE" {
			sessionDeleteHandler(w, r)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	http.HandleFunc("/ws", wsHandler)
	go handleMessages() // optional, falls Broadcast benötigt
	fmt.Println("Server started on :8080 with session management endpoints:")
	fmt.Println("  POST /api/session/new - Create new session")
	fmt.Println("  POST /api/session/message - Send message to session")
	fmt.Println("  GET /api/session/{id}/history - Get session history")
	fmt.Println("  DELETE /api/session/{id} - Delete session")
	fmt.Println("  WebSocket /ws - Enhanced WebSocket with session support")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
