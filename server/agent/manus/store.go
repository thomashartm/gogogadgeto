package manus

import "context"

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{m: make(map[string][]byte)}
}

type InMemoryStore struct {
	m map[string][]byte
}

func (i *InMemoryStore) Get(ctx context.Context, checkPointID string) ([]byte, bool, error) {
	data, ok := i.m[checkPointID]
	return data, ok, nil
}

func (i *InMemoryStore) Set(ctx context.Context, checkPointID string, checkPoint []byte) error {
	i.m[checkPointID] = checkPoint
	return nil
}
