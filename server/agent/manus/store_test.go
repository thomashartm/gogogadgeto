package manus

import (
	"context"
	"fmt"
	"testing"
)

// Test NewInMemoryStore functionality without external dependencies
func TestNewInMemoryStore_Standalone(t *testing.T) {
	store := NewInMemoryStore()

	if store == nil {
		t.Error("NewInMemoryStore() returned nil")
	}
	if store.m == nil {
		t.Error("NewInMemoryStore() returned store with nil map")
	}
	if len(store.m) != 0 {
		t.Errorf("NewInMemoryStore() returned store with non-empty map, got %d items", len(store.m))
	}
}

func TestInMemoryStore_SetAndGet_Standalone(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	checkpointID := "test-checkpoint-1"
	testData := []byte("test checkpoint data")

	// Test Set
	err := store.Set(ctx, checkpointID, testData)
	if err != nil {
		t.Errorf("Set() returned error: %v", err)
	}

	// Test Get existing data
	data, ok, err := store.Get(ctx, checkpointID)
	if err != nil {
		t.Errorf("Get() returned error: %v", err)
	}
	if !ok {
		t.Error("Get() returned ok=false for existing data")
	}
	if string(data) != string(testData) {
		t.Errorf("Get() returned wrong data, expected %s, got %s", string(testData), string(data))
	}

	// Test Get non-existing data
	data, ok, err = store.Get(ctx, "non-existing-id")
	if err != nil {
		t.Errorf("Get() returned error for non-existing key: %v", err)
	}
	if ok {
		t.Error("Get() returned ok=true for non-existing data")
	}
	if data != nil {
		t.Error("Get() returned non-nil data for non-existing key")
	}
}

func TestInMemoryStore_MultipleOperations_Standalone(t *testing.T) {
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
		if err != nil {
			t.Errorf("Set() returned error for key %s: %v", id, err)
		}
	}

	// Verify all checkpoints can be retrieved
	for id, expectedData := range checkpoints {
		data, ok, err := store.Get(ctx, id)
		if err != nil {
			t.Errorf("Get() returned error for key %s: %v", id, err)
		}
		if !ok {
			t.Errorf("Get() returned ok=false for existing key %s", id)
		}
		if string(data) != string(expectedData) {
			t.Errorf("Get() returned wrong data for key %s, expected %s, got %s", id, string(expectedData), string(data))
		}
	}

	// Test overwriting existing checkpoint
	newData := []byte("updated data 1")
	err := store.Set(ctx, "checkpoint-1", newData)
	if err != nil {
		t.Errorf("Set() returned error when overwriting: %v", err)
	}

	data, ok, err := store.Get(ctx, "checkpoint-1")
	if err != nil {
		t.Errorf("Get() returned error after overwrite: %v", err)
	}
	if !ok {
		t.Error("Get() returned ok=false after overwrite")
	}
	if string(data) != string(newData) {
		t.Errorf("Get() returned wrong data after overwrite, expected %s, got %s", string(newData), string(data))
	}
}

func TestInMemoryStore_EdgeCases_Standalone(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryStore()

	// Test with empty string key
	err := store.Set(ctx, "", []byte("data"))
	if err != nil {
		t.Errorf("Set() returned error for empty key: %v", err)
	}

	data, ok, err := store.Get(ctx, "")
	if err != nil {
		t.Errorf("Get() returned error for empty key: %v", err)
	}
	if !ok {
		t.Error("Get() returned ok=false for empty key")
	}
	if string(data) != "data" {
		t.Errorf("Get() returned wrong data for empty key, expected 'data', got %s", string(data))
	}

	// Test with nil data
	err = store.Set(ctx, "nil-test", nil)
	if err != nil {
		t.Errorf("Set() returned error for nil data: %v", err)
	}

	data, ok, err = store.Get(ctx, "nil-test")
	if err != nil {
		t.Errorf("Get() returned error for nil data: %v", err)
	}
	if !ok {
		t.Error("Get() returned ok=false for nil data")
	}
	if data != nil {
		t.Error("Get() returned non-nil data when nil was stored")
	}

	// Test with empty data
	err = store.Set(ctx, "empty-test", []byte{})
	if err != nil {
		t.Errorf("Set() returned error for empty data: %v", err)
	}

	data, ok, err = store.Get(ctx, "empty-test")
	if err != nil {
		t.Errorf("Get() returned error for empty data: %v", err)
	}
	if !ok {
		t.Error("Get() returned ok=false for empty data")
	}
	if len(data) != 0 {
		t.Errorf("Get() returned wrong data for empty bytes, expected length 0, got %d", len(data))
	}
}

// Test NodeKey constants (they're simple string constants)
func TestNodeKeyConstants_Standalone(t *testing.T) {
	expectedKeys := map[string]string{
		"Human":           NodeKeyHuman,
		"InputConverter":  NodeKeyInputConvert,
		"ChatModel":       NodeKeyChatModel,
		"ToolsNode":       NodeKeyToolsNode,
		"OutputConverter": NodeKeyOutputConvert,
	}

	for expected, actual := range expectedKeys {
		if actual != expected {
			t.Errorf("NodeKey mismatch: expected %s, got %s", expected, actual)
		}
	}

	// Test uniqueness
	keys := []string{
		NodeKeyHuman,
		NodeKeyInputConvert,
		NodeKeyChatModel,
		NodeKeyToolsNode,
		NodeKeyOutputConvert,
	}

	keySet := make(map[string]bool)
	for _, key := range keys {
		if keySet[key] {
			t.Errorf("Duplicate node key found: %s", key)
		}
		keySet[key] = true
	}

	if len(keySet) != 5 {
		t.Errorf("Expected 5 unique node keys, got %d", len(keySet))
	}
}

// Benchmark tests
func BenchmarkInMemoryStore_Set_Standalone(b *testing.B) {
	ctx := context.Background()
	store := NewInMemoryStore()
	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Set(ctx, fmt.Sprintf("key-%d", i), testData)
	}
}

func BenchmarkInMemoryStore_Get_Standalone(b *testing.B) {
	ctx := context.Background()
	store := NewInMemoryStore()
	testData := []byte("benchmark test data")

	// Setup data
	for i := 0; i < 100; i++ {
		store.Set(ctx, fmt.Sprintf("key-%d", i), testData)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(ctx, fmt.Sprintf("key-%d", i%100))
	}
}

func BenchmarkInMemoryStore_SetGet_Standalone(b *testing.B) {
	ctx := context.Background()
	store := NewInMemoryStore()
	testData := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		store.Set(ctx, key, testData)
		store.Get(ctx, key)
	}
}
