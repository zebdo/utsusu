package storage

import (
	"errors"
	"sync"

	"github.com/zebdo/utsusu/internal/structs"
)

type memoryStore struct {
	mu      sync.RWMutex
	threads map[string]structs.Thread // key: thread.ID
}

func NewMemory() Storage {
	return &memoryStore{ threads: make(map[string]structs.Thread) }
}

func (m *memoryStore) SaveThread(t structs.Thread) error {
	m.mu.Lock(); defer m.mu.Unlock()
	m.threads[t.ID] = t
	return nil
}

func (m *memoryStore) GetThread(id string) (*structs.Thread, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	th, ok := m.threads[id]
	if !ok { return nil, errors.New("not found") }
	return &th, nil
}

func (m *memoryStore) ListThreads(board string) ([]structs.Thread, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	res := make([]structs.Thread, 0)
	for _, th := range m.threads {
		if th.Board == board {
			res = append(res, th)
		}
	}
	return res, nil
}
