package outbound

import (
	"context"
	"sync"
	"time"
)

type MockLockService struct {
	mu    sync.Mutex
	locks map[string]bool
}

func NewMockLockService() *MockLockService {
	return &MockLockService{
		locks: make(map[string]bool),
	}
}

func (m *MockLockService) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.locks[key] {
		return false, nil
	}
	m.locks[key] = true
	return true, nil
}

func (m *MockLockService) ReleaseLock(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, key)
	return nil
}

// Compile-time check
var _ LockService = (*MockLockService)(nil)
