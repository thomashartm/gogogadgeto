package manus

import (
	"fmt"
	"testing"

	"github.com/cloudwego/eino/schema"
)

// Test extractLastMessage function
func TestExtractLastMessage_Standalone(t *testing.T) {
	tests := []struct {
		name     string
		input    []*schema.Message
		expected string
	}{
		{
			name:     "empty message slice",
			input:    []*schema.Message{},
			expected: "",
		},
		{
			name: "single message",
			input: []*schema.Message{
				{Content: "Hello world", Role: schema.User},
			},
			expected: "Hello world",
		},
		{
			name: "multiple messages - returns last",
			input: []*schema.Message{
				{Content: "First message", Role: schema.User},
				{Content: "Second message", Role: schema.Assistant},
				{Content: "Last message", Role: schema.User},
			},
			expected: "Last message",
		},
		{
			name: "message with empty content",
			input: []*schema.Message{
				{Content: "First message", Role: schema.User},
				{Content: "", Role: schema.Assistant},
			},
			expected: "",
		},
		{
			name: "single message with long content",
			input: []*schema.Message{
				{Content: "This is a very long message content that spans multiple lines and contains various characters and symbols !@#$%^&*()", Role: schema.Assistant},
			},
			expected: "This is a very long message content that spans multiple lines and contains various characters and symbols !@#$%^&*()",
		},
		{
			name: "message with special characters",
			input: []*schema.Message{
				{Content: "Message with 中文 and emojis 🚀🎉", Role: schema.User},
			},
			expected: "Message with 中文 and emojis 🚀🎉",
		},
		{
			name: "message with newlines and tabs",
			input: []*schema.Message{
				{Content: "Line 1\nLine 2\n\tTabbed line", Role: schema.Assistant},
			},
			expected: "Line 1\nLine 2\n\tTabbed line",
		},
		{
			name: "message with JSON-like content",
			input: []*schema.Message{
				{Content: `{"key": "value", "number": 42}`, Role: schema.Assistant},
			},
			expected: `{"key": "value", "number": 42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLastMessage(tt.input)
			if result != tt.expected {
				t.Errorf("extractLastMessage() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

// Test with different message roles using predefined constants
func TestExtractLastMessage_Roles_Standalone(t *testing.T) {
	tests := []struct {
		name     string
		message  *schema.Message
		expected string
	}{
		{
			name:     "user message",
			message:  &schema.Message{Content: "User input", Role: schema.User},
			expected: "User input",
		},
		{
			name:     "assistant message",
			message:  &schema.Message{Content: "Assistant response", Role: schema.Assistant},
			expected: "Assistant response",
		},
		{
			name:     "system message",
			message:  &schema.Message{Content: "System instruction", Role: schema.System},
			expected: "System instruction",
		},
		{
			name:     "tool message",
			message:  &schema.Message{Content: "Tool result", Role: schema.Tool},
			expected: "Tool result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []*schema.Message{tt.message}
			result := extractLastMessage(input)
			if result != tt.expected {
				t.Errorf("extractLastMessage() with role %v = %q, expected %q", tt.message.Role, result, tt.expected)
			}
		})
	}
}

// Test performance with large message arrays
func TestExtractLastMessage_Performance_Standalone(t *testing.T) {
	// Create a large slice of messages
	messages := make([]*schema.Message, 10000)
	for i := 0; i < 10000; i++ {
		messages[i] = &schema.Message{
			Content: "Message content",
			Role:    schema.User,
		}
	}
	// Set the last message to have unique content
	messages[9999].Content = "Last message content"

	result := extractLastMessage(messages)
	if result != "Last message content" {
		t.Errorf("extractLastMessage() with large array = %q, expected %q", result, "Last message content")
	}
}

// Benchmark tests for extractLastMessage
func BenchmarkExtractLastMessage_Standalone(b *testing.B) {
	messages := make([]*schema.Message, 100)
	for i := 0; i < 100; i++ {
		messages[i] = &schema.Message{
			Content: "Test message content",
			Role:    schema.User,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractLastMessage(messages)
	}
}

func BenchmarkExtractLastMessage_VaryingSizes_Standalone(b *testing.B) {
	sizes := []int{1, 10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size-%d", size), func(b *testing.B) {
			messages := make([]*schema.Message, size)
			for i := 0; i < size; i++ {
				messages[i] = &schema.Message{
					Content: "Test message content",
					Role:    schema.User,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				extractLastMessage(messages)
			}
		})
	}
}
