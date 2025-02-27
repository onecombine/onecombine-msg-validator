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

func (ms *MemoryStore) Keys() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	keys := make([]string, 0)

	for k, _ := range ms.store {
		keys = append(keys, k)
	}

	return keys
}

func (ms *MemoryStore) GetAll() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.store
}

func (ms *MemoryStore) Delete(key string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.store, key)
}
