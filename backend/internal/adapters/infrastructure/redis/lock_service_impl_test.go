package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pos-backend/internal/adapters/infrastructure/redis"
	redisclient "github.com/redis/go-redis/v9"
)

func TestRedisLockService_AcquireAndRelease(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redisclient.NewClient(&redisclient.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	lockService := redis.NewRedisLockService(rdb)
	ctx := context.Background()

	lockKey := "lock:checkout:test-key-100"

	// Client A acquires lock
	acquired, err := lockService.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error acquiring lock, got %v", err)
	}
	if !acquired {
		t.Errorf("expected lock acquisition to succeed")
	}

	// Client B tries to acquire lock on same key -> should fail (return false)
	acquiredSecond, err := lockService.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error checking locked key, got %v", err)
	}
	if acquiredSecond {
		t.Errorf("expected lock acquisition for Client B to fail while Client A holds lock")
	}

	// Client A releases lock
	err = lockService.ReleaseLock(ctx, lockKey)
	if err != nil {
		t.Fatalf("expected no error releasing lock, got %v", err)
	}

	// Client B tries again after release -> should succeed
	acquiredThird, err := lockService.AcquireLock(ctx, lockKey, 5*time.Second)
	if err != nil {
		t.Fatalf("expected no error acquiring lock after release, got %v", err)
	}
	if !acquiredThird {
		t.Errorf("expected lock acquisition to succeed after release")
	}
}

func TestRedisLockService_TTLExpiration(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redisclient.NewClient(&redisclient.Options{
		Addr: mr.Addr(),
	})
	defer rdb.Close()

	lockService := redis.NewRedisLockService(rdb)
	ctx := context.Background()

	lockKey := "lock:checkout:ttl-test"

	// Acquire lock with 1 second TTL
	acquired, err := lockService.AcquireLock(ctx, lockKey, 1*time.Second)
	if err != nil {
		t.Fatalf("expected no error acquiring lock, got %v", err)
	}
	if !acquired {
		t.Fatalf("expected lock acquisition to succeed")
	}

	// Fast-forward miniredis clock by 2 seconds
	mr.FastForward(2 * time.Second)

	// Try acquiring lock again after TTL expired
	acquiredAfterTTL, err := lockService.AcquireLock(ctx, lockKey, 1*time.Second)
	if err != nil {
		t.Fatalf("expected no error acquiring lock after TTL, got %v", err)
	}
	if !acquiredAfterTTL {
		t.Errorf("expected lock acquisition to succeed after TTL expired")
	}
}
