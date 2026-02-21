package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// ClosedPRTTL PR 關閉後保留 7 天
	ClosedPRTTL = 7 * 24 * time.Hour
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStore 建立 Redis storage（接受 redis:// URL）
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	client := redis.NewClient(opts)

	ctx := context.Background()

	// 測試連線
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// Set 儲存 PR → Thread 對應，不設定 TTL（永久保存）
func (r *RedisStore) Set(prID, threadID string) error {
	// TTL = 0 表示永不過期
	if err := r.client.Set(r.ctx, prID, threadID, 0).Err(); err != nil {
		return fmt.Errorf("failed to set mapping: %w", err)
	}
	return nil
}

// Get 取得 Thread ID
func (r *RedisStore) Get(prID string) (string, bool, error) {
	val, err := r.client.Get(r.ctx, prID).Result()

	// Key 不存在
	if err == redis.Nil {
		return "", false, nil
	}

	// 其他錯誤
	if err != nil {
		return "", false, fmt.Errorf("failed to get mapping: %w", err)
	}

	return val, true, nil
}

// Delete 刪除對應關係
func (r *RedisStore) Delete(prID string) error {
	if err := r.client.Del(r.ctx, prID).Err(); err != nil {
		return fmt.Errorf("failed to delete mapping: %w", err)
	}
	return nil
}

// MarkAsClosed PR 關閉時呼叫，設定 7 天 TTL
func (r *RedisStore) MarkAsClosed(prID string) error {
	// 先取得現有的 threadID
	threadID, exists, err := r.Get(prID)
	if err != nil {
		return err
	}

	// 如果 mapping 不存在，不做事（可能已被刪除）
	if !exists {
		return nil
	}

	// 重新設定，帶 7 天 TTL
	if err := r.client.Set(r.ctx, prID, threadID, ClosedPRTTL).Err(); err != nil {
		return fmt.Errorf("failed to mark as closed: %w", err)
	}

	return nil
}

// Close 關閉 Redis 連線
func (r *RedisStore) Close() error {
	return r.client.Close()
}
