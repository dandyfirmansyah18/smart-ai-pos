package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/pos-backend/internal/ports/outbound"
	"github.com/redis/go-redis/v9"
)

type RedisLockService struct {
	client *redis.Client
}

func NewRedisLockService(client *redis.Client) *RedisLockService {
	return &RedisLockService{client: client}
}

func (s *RedisLockService) AcquireLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	success, err := s.client.SetNX(ctx, key, "locked", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock for key %s: %w", key, err)
	}
	return success, nil
}

func (s *RedisLockService) ReleaseLock(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to release lock for key %s: %w", key, err)
	}
	return nil
}

// Compile-time check that RedisLockService implements outbound.LockService
var _ outbound.LockService = (*RedisLockService)(nil)
