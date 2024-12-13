package partners

import (
	"errors"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	store map[string]interface{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		store: make(map[string]interface{}),
	}
}

func (ms *MemoryStore) Set(key string, value interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.store[key] = value
}

func (ms *MemoryStore) Get(key string) (interface{}, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	value, ok := ms.store[key]
	if !ok {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (ms *MemoryStore) Delete(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.store, key)
}
