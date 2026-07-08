// Package cache provides Redis-backed distributed cache and session state for ACSGO.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	deviceCacheTTL  = 5 * time.Minute
	sessionCacheTTL = 30 * time.Minute
	lockTTL         = 30 * time.Second
)

// Client wraps redis.Client with ACS-specific helpers.
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client.
func NewClient(url, password string, db int) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// fallback: treat url as address
		opts = &redis.Options{
			Addr:     url,
			Password: password,
			DB:       db,
		}
	}
	opts.Password = password
	opts.DB = db

	rdb := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Close shuts down the Redis connection.
func (c *Client) Close() error {
	return c.rdb.Close()
}

// ---- Device Cache ----

// SetDevice caches device data.
func (c *Client) SetDevice(ctx context.Context, deviceID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, deviceKey(deviceID), b, deviceCacheTTL).Err()
}

// GetDevice retrieves cached device data into dst.
func (c *Client) GetDevice(ctx context.Context, deviceID string, dst interface{}) error {
	b, err := c.rdb.Get(ctx, deviceKey(deviceID)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// DeleteDevice removes a device from cache.
func (c *Client) DeleteDevice(ctx context.Context, deviceID string) error {
	return c.rdb.Del(ctx, deviceKey(deviceID)).Err()
}

// ---- Session State ----

// SetSession stores session data keyed by device ID.
func (c *Client) SetSession(ctx context.Context, deviceID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, sessionKey(deviceID), b, sessionCacheTTL).Err()
}

// GetSession retrieves session data for a device.
func (c *Client) GetSession(ctx context.Context, deviceID string, dst interface{}) error {
	b, err := c.rdb.Get(ctx, sessionKey(deviceID)).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

// DeleteSession removes a session.
func (c *Client) DeleteSession(ctx context.Context, deviceID string) error {
	return c.rdb.Del(ctx, sessionKey(deviceID)).Err()
}

// ---- Online Tracking ----

// MarkOnline records a device as online with TTL.
func (c *Client) MarkOnline(ctx context.Context, deviceID string, ttl time.Duration) error {
	return c.rdb.Set(ctx, onlineKey(deviceID), "1", ttl).Err()
}

// IsOnline checks whether a device is marked as online.
func (c *Client) IsOnline(ctx context.Context, deviceID string) bool {
	v, err := c.rdb.Exists(ctx, onlineKey(deviceID)).Result()
	return err == nil && v > 0
}

// ---- Distributed Locks ----

// AcquireLock attempts to acquire a distributed lock.
// Returns true if lock was acquired.
func (c *Client) AcquireLock(ctx context.Context, resource string) (bool, error) {
	ok, err := c.rdb.SetNX(ctx, lockKey(resource), "1", lockTTL).Result()
	return ok, err
}

// ReleaseLock releases a distributed lock.
func (c *Client) ReleaseLock(ctx context.Context, resource string) error {
	return c.rdb.Del(ctx, lockKey(resource)).Err()
}

// ---- Pub/Sub for real-time events ----

// Publish sends an event to a channel.
func (c *Client) Publish(ctx context.Context, channel string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, channel, b).Err()
}

// Subscribe subscribes to a channel and calls handler for each message.
func (c *Client) Subscribe(ctx context.Context, channel string, handler func(payload []byte)) error {
	sub := c.rdb.Subscribe(ctx, channel)
	ch := sub.Channel()
	go func() {
		for msg := range ch {
			handler([]byte(msg.Payload))
		}
	}()
	return nil
}

// ---- Counter helpers (for rate limiting / metrics) ----

// Incr atomically increments a counter and returns the new value.
func (c *Client) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// ---- Key helpers ----
func deviceKey(id string) string  { return "device:" + id }
func sessionKey(id string) string { return "session:" + id }
func onlineKey(id string) string  { return "online:" + id }
func lockKey(r string) string     { return "lock:" + r }
