package manus

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test extractLastMessage function without any external dependencies
func TestExtractLastMessage_Core(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLastMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test InMemoryStore functionality
func TestNewInMemoryStore_Core(t *testing.T) {
	store := NewInMemoryStore()

	assert.NotNil(t, store)
	assert.NotNil(t, store.m)
	assert.Equal(t, 0, len(store.m))
}

func TestInMemoryStore_SetAndGet_Core(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	checkpointID := "test-checkpoint-1"
	testData := []byte("test checkpoint data")

	// Test Set
	err := store.Set(ctx, checkpointID, testData)
	require.NoError(t, err)

	// Test Get existing data
	data, ok, err := store.Get(ctx, checkpointID)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, testData, data)

	// Test Get non-existing data
	data, ok, err = store.Get(ctx, "non-existing-id")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, data)
}

func TestInMemoryStore_MultipleOperations_Core(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	// Test storing multiple checkpoints
	checkpoints := map[string][]byte{
		"checkpoint-1": []byte("data 1"),
		"checkpoint-2": []byte("data 2"),
		"checkpoint-3": []byte("data 3"),
	}

	// Set all checkpoints
	for id, data := range checkpoints {
		err := store.Set(ctx, id, data)
		require.NoError(t, err)
	}

	// Verify all checkpoints can be retrieved
	for id, expectedData := range checkpoints {
		data, ok, err := store.Get(ctx, id)
		require.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, expectedData, data)
	}

	// Test overwriting existing checkpoint
	newData := []byte("updated data 1")
	err := store.Set(ctx, "checkpoint-1", newData)
	require.NoError(t, err)

	data, ok, err := store.Get(ctx, "checkpoint-1")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, newData, data)
}

func TestInMemoryStore_EdgeCases_Core(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	// Test with empty string key
	err := store.Set(ctx, "", []byte("data"))
	assert.NoError(t, err)

	data, ok, err := store.Get(ctx, "")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("data"), data)

	// Test with nil data
	err = store.Set(ctx, "nil-test", nil)
	assert.NoError(t, err)

	data, ok, err = store.Get(ctx, "nil-test")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Nil(t, data)

	// Test with empty data
	err = store.Set(ctx, "empty-test", []byte{})
	assert.NoError(t, err)

	data, ok, err = store.Get(ctx, "empty-test")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte{}, data)

	// Test with very long key
	longKey := string(make([]byte, 1000))
	err = store.Set(ctx, longKey, []byte("long key data"))
	assert.NoError(t, err)

	data, ok, err = store.Get(ctx, longKey)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("long key data"), data)
}

func TestInMemoryStore_ConcurrentAccess_Core(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	// Note: This is a basic test. For true concurrent testing,
	// we'd need goroutines and sync mechanisms

	// Simulate multiple operations
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		data := []byte(fmt.Sprintf("data-%d", i))

		err := store.Set(ctx, key, data)
		assert.NoError(t, err)
	}

	// Verify all data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		expectedData := []byte(fmt.Sprintf("data-%d", i))

		data, ok, err := store.Get(ctx, key)
		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, expectedData, data)
	}
}

// Test the node key constants
func TestNodeKeyConstants_Core(t *testing.T) {
	expectedKeys := map[string]string{
		"NodeKeyHuman":         "Human",
		"NodeKeyInputConvert":  "InputConverter",
		"NodeKeyChatModel":     "ChatModel",
		"NodeKeyToolsNode":     "ToolsNode",
		"NodeKeyOutputConvert": "OutputConverter",
	}

	assert.Equal(t, expectedKeys["NodeKeyHuman"], NodeKeyHuman)
	assert.Equal(t, expectedKeys["NodeKeyInputConvert"], NodeKeyInputConvert)
	assert.Equal(t, expectedKeys["NodeKeyChatModel"], NodeKeyChatModel)
	assert.Equal(t, expectedKeys["NodeKeyToolsNode"], NodeKeyToolsNode)
	assert.Equal(t, expectedKeys["NodeKeyOutputConvert"], NodeKeyOutputConvert)
}

func TestNodeKeyConstants_Uniqueness_Core(t *testing.T) {
	// Ensure all node keys are unique
	keys := []string{
		NodeKeyHuman,
		NodeKeyInputConvert,
		NodeKeyChatModel,
		NodeKeyToolsNode,
		NodeKeyOutputConvert,
	}

	keySet := make(map[string]bool)
	for _, key := range keys {
		assert.False(t, keySet[key], "Duplicate node key found: %s", key)
		keySet[key] = true
	}

	assert.Len(t, keySet, 5, "Should have exactly 5 unique node keys")
}

func TestNodeKeyConstants_NonEmpty_Core(t *testing.T) {
	// Ensure all node keys are non-empty strings
	keys := []string{
		NodeKeyHuman,
		NodeKeyInputConvert,
		NodeKeyChatModel,
		NodeKeyToolsNode,
		NodeKeyOutputConvert,
	}

	for _, key := range keys {
		assert.NotEmpty(t, key, "Node key should not be empty")
		assert.NotContains(t, key, " ", "Node key should not contain spaces")
	}
}

// Benchmark tests
func BenchmarkExtractLastMessage_Core(b *testing.B) {
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

func BenchmarkExtractLastMessage_VaryingSizes_Core(b *testing.B) {
	sizes := []int{1, 10, 100, 1000}

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

func BenchmarkInMemoryStore_Core(b *testing.B) {
	ctx := context.Background()
	store := NewInMemoryStore()
	testData := []byte("benchmark test data")

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			store.Set(ctx, fmt.Sprintf("key-%d", i), testData)
		}
	})

	b.Run("Get", func(b *testing.B) {
		// Setup data
		for i := 0; i < 100; i++ {
			store.Set(ctx, fmt.Sprintf("key-%d", i), testData)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			store.Get(ctx, fmt.Sprintf("key-%d", i%100))
		}
	})

	b.Run("SetGet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i)
			store.Set(ctx, key, testData)
			store.Get(ctx, key)
		}
	})
}
