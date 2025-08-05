# Tracer Module Documentation

The tracer module provides comprehensive logging capabilities for node execution in the Eino framework, designed to integrate seamlessly with `WithStatePreHandler` and `WithStatePostHandler`.

## Features

### 🔍 **NodeTracer**
- **Intelligent Type Detection**: Automatically handles different input/output types
- **Schema Message Support**: Special handling for `[]*schema.Message` and `*schema.Message`
- **JSON Formatting**: Pretty-prints complex objects
- **Content Truncation**: Prevents log spam with configurable limits
- **Enable/Disable**: Can be turned on/off per tracer

### ⏱️ **ExecutionTimer**
- **Timing Integration**: Tracks execution duration for nodes
- **Automatic Logging**: Logs start and end with duration
- **Tracer Integration**: Works with NodeTracer for unified logging

### ⚙️ **TracerConfig**
- **Global Configuration**: Control all tracers from one place
- **Selective Enabling**: Enable/disable pre-handlers, post-handlers, timers, etc.
- **Performance Tuning**: Adjust log length limits and truncation

## Usage Examples

### Basic Integration

```go
import "gogogajeto/util"

// Create a tracer for your node
tracer := util.NewNodeTracer("MyCustomNode")

// In your node handler
func myNodeHandler(ctx context.Context, input string) (string, error) {
    tracer.SimpleTracePreHandler(ctx, input)
    
    // Your business logic
    result := processInput(input)
    
    tracer.SimpleTracePostHandler(ctx, result)
    return result, nil
}
```

### Integration with Eino Framework

```go
// In your compose.AddLambdaNode
err := g.AddLambdaNode(NodeKeyExample, compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
    tracer := util.NewNodeTracer("ExampleNode")
    tracer.SimpleTracePreHandler(ctx, input)
    
    // Your processing logic
    result := "processed: " + input
    
    tracer.SimpleTracePostHandler(ctx, result)
    return result, nil
}))

// In your WithStatePreHandler/WithStatePostHandler
compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, state *State) ([]*schema.Message, error) {
    tracer := util.NewNodeTracer("ChatModel")
    tracer.SimpleTracePreHandler(ctx, in)
    
    // Your pre-handler logic
    
    return processedMessages, nil
})
```

### Configuration

```go
// Configure global tracer behavior
config := &util.TracerConfig{
    EnablePreHandlers:  true,
    EnablePostHandlers: true,
    EnableTimers:       true,
    EnableStateChanges: false,  // Can be verbose
    MaxLogLength:       500,
    TruncateMessages:   true,
}
util.SetGlobalTracerConfig(config)
```

### Execution Timing

```go
tracer := util.NewNodeTracer("TimedNode")
timer := util.NewExecutionTimer("TimedNode", tracer)

timer.Start()
// ... do work ...
duration := timer.End()

// Automatically logs: "=== TimedNode EXECUTION START/END ==="
```

## Output Examples

### String Input/Output
```
=== MyNode Node PRE-HANDLER ===
Input type: string
Input string length: 25
Input: What is the weather today?

=== MyNode Node POST-HANDLER ===
Output type: string  
Output string length: 42
Output: The weather is sunny with 25°C temperature
=== MyNode Node END ===
```

### Schema Messages
```
=== ChatModel Node PRE-HANDLER ===
Input type: []*schema.Message
Input messages count: 2
Message[0] Role: system
Message[0] Content length: 156
Message[1] Role: user  
Message[1] Content length: 25

=== ChatModel Node POST-HANDLER ===
Output type: *schema.Message
Output message role: assistant
Output content length: 42
Tool calls count: 1
Tool call 0: kali_info_gathering
Tool call 0 ID: call_xyz123
Tool arguments: {"tool":"whois","target":"example.com"}
=== ChatModel Node END ===
```

### JSON Objects
```
=== CustomNode Node PRE-HANDLER ===
Input type: map[string]interface {}
Input JSON: {"command":"scan","target":"192.168.1.1","options":["fast","verbose"]}

=== CustomNode Node POST-HANDLER ===  
Output JSON: {"status":"completed","results":["port 22 open","port 80 open"],"duration":"2.3s"}
=== CustomNode Node END ===
```

## Integration with Existing Code

The tracer module is designed to enhance existing logging without disrupting current functionality:

1. **Non-intrusive**: Doesn't change existing API contracts
2. **Additive**: Provides additional structured logging alongside existing logs
3. **Configurable**: Can be disabled or customized as needed
4. **Type-aware**: Handles different data types intelligently

## Performance Considerations

- **Conditional Logging**: Only processes when tracer is enabled
- **Truncation**: Prevents large objects from overwhelming logs
- **Lazy Evaluation**: JSON serialization only when needed
- **Global Config**: Single point to disable all tracing for production

## Best Practices

1. **Create tracers per node**: `util.NewNodeTracer("NodeName")`
2. **Use Simple methods**: `SimpleTracePreHandler` and `SimpleTracePostHandler` 
3. **Configure globally**: Set limits and behavior via `TracerConfig`
4. **Monitor performance**: Disable in production if log volume becomes an issue
5. **Structured naming**: Use consistent node names for easier log analysis 