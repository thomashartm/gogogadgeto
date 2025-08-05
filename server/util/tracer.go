package util

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// NodeTracer provides consistent logging for node execution
type NodeTracer struct {
	NodeName string
	Enabled  bool
}

// NewNodeTracer creates a new tracer for a specific node
func NewNodeTracer(nodeName string) *NodeTracer {
	return &NodeTracer{
		NodeName: nodeName,
		Enabled:  true,
	}
}

// TracePreHandler logs information before node execution
func (t *NodeTracer) TracePreHandler(ctx context.Context, path compose.NodePath, input interface{}) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s Node PRE-HANDLER ===", t.NodeName))
	LogMessage(fmt.Sprintf("Node path: %v", path))

	// Log input details based on type
	t.logInputDetails(input)
}

// TracePostHandler logs information after node execution
func (t *NodeTracer) TracePostHandler(ctx context.Context, path compose.NodePath, output interface{}) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s Node POST-HANDLER ===", t.NodeName))
	LogMessage(fmt.Sprintf("Node path: %v", path))

	// Log output details based on type
	t.logOutputDetails(output)

	LogMessage(fmt.Sprintf("=== %s Node END ===", t.NodeName))
}

// TraceError logs error information
func (t *NodeTracer) TraceError(ctx context.Context, path compose.NodePath, err error) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s Node ERROR ===", t.NodeName))
	LogMessage(fmt.Sprintf("Node path: %v", path))
	LogMessage(fmt.Sprintf("Error: %v", err))
}

// TraceStateChange logs state modifications
func (t *NodeTracer) TraceStateChange(ctx context.Context, path compose.NodePath, stateBefore, stateAfter interface{}) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s State Change ===", t.NodeName))
	t.logStateComparison(stateBefore, stateAfter)
}

// SimpleTracePreHandler logs information before node execution without requiring NodePath
func (t *NodeTracer) SimpleTracePreHandler(ctx context.Context, input interface{}) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s Node PRE-HANDLER ===", t.NodeName))

	// Log input details based on type
	t.logInputDetails(input)
}

// SimpleTracePostHandler logs information after node execution without requiring NodePath
func (t *NodeTracer) SimpleTracePostHandler(ctx context.Context, output interface{}) {
	if !t.Enabled {
		return
	}

	LogMessage(fmt.Sprintf("=== %s Node POST-HANDLER ===", t.NodeName))

	// Log output details based on type
	t.logOutputDetails(output)

	LogMessage(fmt.Sprintf("=== %s Node END ===", t.NodeName))
}

// logInputDetails logs details about the input based on its type
func (t *NodeTracer) logInputDetails(input interface{}) {
	if input == nil {
		LogMessage("Input: nil")
		return
	}

	inputType := reflect.TypeOf(input).String()
	LogMessage(fmt.Sprintf("Input type: %s", inputType))

	switch v := input.(type) {
	case string:
		LogMessage(fmt.Sprintf("Input string length: %d", len(v)))
		if len(v) <= 200 {
			LogMessage(fmt.Sprintf("Input: %s", v))
		} else {
			LogMessage(fmt.Sprintf("Input (truncated): %s...", v[:200]))
		}

	case []*schema.Message:
		LogMessage(fmt.Sprintf("Input messages count: %d", len(v)))
		for i, msg := range v {
			if i < 5 { // Limit to first 5 messages to avoid spam
				LogMessage(fmt.Sprintf("Message[%d] Role: %s", i, msg.Role))
				LogMessage(fmt.Sprintf("Message[%d] Content length: %d", i, len(msg.Content)))
				if len(msg.ToolCalls) > 0 {
					LogMessage(fmt.Sprintf("Message[%d] Tool calls: %d", i, len(msg.ToolCalls)))
				}
			}
		}
		if len(v) > 5 {
			LogMessage(fmt.Sprintf("... and %d more messages", len(v)-5))
		}

	case *schema.Message:
		LogMessage(fmt.Sprintf("Single message role: %s", v.Role))
		LogMessage(fmt.Sprintf("Content length: %d", len(v.Content)))
		if len(v.ToolCalls) > 0 {
			LogMessage(fmt.Sprintf("Tool calls: %d", len(v.ToolCalls)))
			for i, tc := range v.ToolCalls {
				LogMessage(fmt.Sprintf("Tool call %d: %s", i, tc.Function.Name))
				LogMessage(fmt.Sprintf("Tool call %d ID: %s", i, tc.ID))
				LogMessage(fmt.Sprintf("Tool arguments: %s", tc.Function.Arguments))
			}
		}

	default:
		// Try to serialize as JSON for other types
		if jsonBytes, err := json.Marshal(input); err == nil {
			jsonStr := string(jsonBytes)
			if len(jsonStr) <= 500 {
				LogMessage(fmt.Sprintf("Input JSON: %s", jsonStr))
			} else {
				LogMessage(fmt.Sprintf("Input JSON (truncated): %s...", jsonStr[:500]))
			}
		} else {
			LogMessage(fmt.Sprintf("Input: %+v", input))
		}
	}
}

// logOutputDetails logs details about the output based on its type
func (t *NodeTracer) logOutputDetails(output interface{}) {
	if output == nil {
		LogMessage("Output: nil")
		return
	}

	outputType := reflect.TypeOf(output).String()
	LogMessage(fmt.Sprintf("Output type: %s", outputType))

	switch v := output.(type) {
	case string:
		LogMessage(fmt.Sprintf("Output string length: %d", len(v)))
		if len(v) <= 200 {
			LogMessage(fmt.Sprintf("Output: %s", v))
		} else {
			LogMessage(fmt.Sprintf("Output (truncated): %s...", v[:200]))
		}

	case []*schema.Message:
		LogMessage(fmt.Sprintf("Output messages count: %d", len(v)))
		for i, msg := range v {
			if i < 3 { // Limit to first 3 messages
				LogMessage(fmt.Sprintf("Output[%d] Role: %s", i, msg.Role))
				LogMessage(fmt.Sprintf("Output[%d] Content length: %d", i, len(msg.Content)))
			}
		}

	case *schema.Message:
		LogMessage(fmt.Sprintf("Output message role: %s", v.Role))
		LogMessage(fmt.Sprintf("Output content length: %d", len(v.Content)))
		if len(v.ToolCalls) > 0 {
			LogMessage(fmt.Sprintf("Tool calls count: %d", len(v.ToolCalls)))
			for i, tc := range v.ToolCalls {
				LogMessage(fmt.Sprintf("Tool call %d: %s", i, tc.Function.Name))
				LogMessage(fmt.Sprintf("Tool call %d ID: %s", i, tc.ID))
				LogMessage(fmt.Sprintf("Tool arguments: %s", tc.Function.Arguments))
			}
		}

	default:
		// Try to serialize as JSON for other types
		if jsonBytes, err := json.Marshal(output); err == nil {
			jsonStr := string(jsonBytes)
			if len(jsonStr) <= 500 {
				LogMessage(fmt.Sprintf("Output JSON: %s", jsonStr))
			} else {
				LogMessage(fmt.Sprintf("Output JSON (truncated): %s...", jsonStr[:500]))
			}
		} else {
			LogMessage(fmt.Sprintf("Output: %+v", output))
		}
	}
}

// logStateComparison logs changes between before and after state
func (t *NodeTracer) logStateComparison(before, after interface{}) {
	// This is a simplified state comparison - could be enhanced based on state structure
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)

	if string(beforeJSON) != string(afterJSON) {
		LogMessage("State changed during node execution")
		LogMessage(fmt.Sprintf("State before length: %d bytes", len(beforeJSON)))
		LogMessage(fmt.Sprintf("State after length: %d bytes", len(afterJSON)))
	} else {
		LogMessage("State unchanged during node execution")
	}
}

// ExecutionTimer tracks execution time for nodes
type ExecutionTimer struct {
	NodeName  string
	StartTime time.Time
	tracer    *NodeTracer
}

// NewExecutionTimer creates a new execution timer
func NewExecutionTimer(nodeName string, tracer *NodeTracer) *ExecutionTimer {
	return &ExecutionTimer{
		NodeName:  nodeName,
		StartTime: time.Now(),
		tracer:    tracer,
	}
}

// Start begins timing
func (et *ExecutionTimer) Start() {
	et.StartTime = time.Now()
	if et.tracer != nil && et.tracer.Enabled {
		LogMessage(fmt.Sprintf("=== %s EXECUTION START ===", et.NodeName))
	}
}

// End stops timing and logs duration
func (et *ExecutionTimer) End() time.Duration {
	duration := time.Since(et.StartTime)
	if et.tracer != nil && et.tracer.Enabled {
		LogMessage(fmt.Sprintf("=== %s EXECUTION END ===", et.NodeName))
		LogMessage(fmt.Sprintf("Execution time: %v", duration))
	}
	return duration
}

// TracerConfig holds configuration for tracer behavior
type TracerConfig struct {
	EnablePreHandlers  bool
	EnablePostHandlers bool
	EnableTimers       bool
	EnableStateChanges bool
	MaxLogLength       int
	TruncateMessages   bool
}

// DefaultTracerConfig returns default tracer configuration
func DefaultTracerConfig() *TracerConfig {
	return &TracerConfig{
		EnablePreHandlers:  true,
		EnablePostHandlers: true,
		EnableTimers:       true,
		EnableStateChanges: false, // Can be verbose
		MaxLogLength:       500,
		TruncateMessages:   true,
	}
}

// SetGlobalTracerConfig sets global configuration for all tracers
var globalTracerConfig = DefaultTracerConfig()

func SetGlobalTracerConfig(config *TracerConfig) {
	globalTracerConfig = config
}

func GetGlobalTracerConfig() *TracerConfig {
	return globalTracerConfig
}

// Helper functions for easy integration with existing handlers

// CreatePreHandler creates a pre-handler function for a node
func CreatePreHandler[T any](tracer *NodeTracer) func(context.Context, compose.NodePath, T) error {
	return func(ctx context.Context, path compose.NodePath, input T) error {
		if globalTracerConfig.EnablePreHandlers {
			tracer.TracePreHandler(ctx, path, input)
		}
		return nil
	}
}

// CreatePostHandler creates a post-handler function for a node
func CreatePostHandler[T any](tracer *NodeTracer) func(context.Context, compose.NodePath, T) error {
	return func(ctx context.Context, path compose.NodePath, output T) error {
		if globalTracerConfig.EnablePostHandlers {
			tracer.TracePostHandler(ctx, path, output)
		}
		return nil
	}
}

// CreateTimedHandlers creates both pre and post handlers with timing
func CreateTimedHandlers[TIn, TOut any](nodeName string) (
	func(context.Context, compose.NodePath, TIn) error,
	func(context.Context, compose.NodePath, TOut) error,
) {
	tracer := NewNodeTracer(nodeName)
	var timer *ExecutionTimer

	preHandler := func(ctx context.Context, path compose.NodePath, input TIn) error {
		if globalTracerConfig.EnableTimers {
			timer = NewExecutionTimer(nodeName, tracer)
			timer.Start()
		}
		if globalTracerConfig.EnablePreHandlers {
			tracer.TracePreHandler(ctx, path, input)
		}
		return nil
	}

	postHandler := func(ctx context.Context, path compose.NodePath, output TOut) error {
		if globalTracerConfig.EnablePostHandlers {
			tracer.TracePostHandler(ctx, path, output)
		}
		if globalTracerConfig.EnableTimers && timer != nil {
			timer.End()
		}
		return nil
	}

	return preHandler, postHandler
}
