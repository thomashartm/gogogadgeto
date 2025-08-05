package util

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ExampleUsage demonstrates how to use the tracer module
func ExampleUsage() {
	ctx := context.Background()

	// Example 1: Basic tracer usage
	tracer := NewNodeTracer("ExampleNode")

	// Trace a simple function execution
	input := "test input"
	tracer.SimpleTracePreHandler(ctx, input)

	// ... do some work ...

	output := "test output"
	tracer.SimpleTracePostHandler(ctx, output)

	// Example 2: Using execution timer
	timer := NewExecutionTimer("ExampleNode", tracer)
	timer.Start()

	// ... do some work ...

	duration := timer.End()
	LogMessage("Execution took: " + duration.String())

	// Example 3: Configuring tracer behavior
	config := &TracerConfig{
		EnablePreHandlers:  true,
		EnablePostHandlers: true,
		EnableTimers:       false,
		EnableStateChanges: false,
		MaxLogLength:       300,
		TruncateMessages:   true,
	}
	SetGlobalTracerConfig(config)

	// Example 4: Creating timed handlers (if you want to use the generic approach)
	preHandler, postHandler := CreateTimedHandlers[string, *schema.Message]("CustomNode")

	// Use in compose framework:
	// g.AddLambdaNode("CustomNode", ...,
	//   compose.WithStatePreHandler(preHandler),
	//   compose.WithStatePostHandler(postHandler))

	_ = preHandler
	_ = postHandler
}

// ExampleNodeHandler shows how to integrate tracer in a custom node handler
func ExampleNodeHandler(ctx context.Context, input string) (string, error) {
	tracer := NewNodeTracer("CustomProcessingNode")

	// Log the start of processing
	tracer.SimpleTracePreHandler(ctx, input)

	// Your business logic here
	result := "processed: " + input

	// Log the completion
	tracer.SimpleTracePostHandler(ctx, result)

	return result, nil
}

// ExampleWithErrorHandling shows how to handle errors with tracer
func ExampleWithErrorHandling(ctx context.Context, input interface{}) error {
	tracer := NewNodeTracer("ErrorProneNode")

	tracer.SimpleTracePreHandler(ctx, input)

	defer func() {
		if r := recover(); r != nil {
			tracer.TraceError(ctx, compose.NodePath{}, r.(error))
		}
	}()

	// Simulate some processing that might fail
	if input == nil {
		err := fmt.Errorf("input cannot be nil")
		tracer.TraceError(ctx, compose.NodePath{}, err)
		return err
	}

	tracer.SimpleTracePostHandler(ctx, "success")
	return nil
}

// ExampleStateComparison shows how to trace state changes
func ExampleStateComparison(ctx context.Context) {
	tracer := NewNodeTracer("StateModifyingNode")

	type ExampleState struct {
		Counter int
		Data    string
	}

	beforeState := &ExampleState{Counter: 0, Data: "initial"}

	// ... modify state ...

	afterState := &ExampleState{Counter: 1, Data: "modified"}

	tracer.TraceStateChange(ctx, compose.NodePath{}, beforeState, afterState)
}
