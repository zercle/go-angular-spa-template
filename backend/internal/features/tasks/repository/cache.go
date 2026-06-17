// STUB FEATURE — delete internal/features/tasks to start your project.

package repository

import (
	"context"
	"time"

	valkeygo "github.com/valkey-io/valkey-go"
)

// Cache is the small port the cached repository depends on. Keeping it narrow
// (instead of using valkeygo.Client directly) makes the decorator trivially
// unit-testable with an in-memory fake. All operations are best-effort: a cache
// outage must never fail a request, so errors are swallowed.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Delete(ctx context.Context, key string)
	Incr(ctx context.Context, key string)
}

// ValkeyCache adapts a valkey-go client to the Cache port.
type ValkeyCache struct {
	client valkeygo.Client
}

// NewValkeyCache returns a Cache backed by Valkey.
func NewValkeyCache(client valkeygo.Client) *ValkeyCache {
	return &ValkeyCache{client: client}
}

var _ Cache = (*ValkeyCache)(nil)

// Get returns the cached bytes for key, or ok=false on a miss or any error.
func (c *ValkeyCache) Get(ctx context.Context, key string) ([]byte, bool) {
	resp := c.client.Do(ctx, c.client.B().Get().Key(key).Build())
	if resp.Error() != nil {
		return nil, false // valkey-nil (miss) or any error → treat as miss
	}
	b, err := resp.AsBytes()
	if err != nil {
		return nil, false
	}
	return b, true
}

// Set stores value under key with the given TTL (best-effort).
func (c *ValkeyCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) {
	_ = c.client.Do(ctx, c.client.B().Set().Key(key).Value(string(value)).ExSeconds(int64(ttl.Seconds())).Build()).Error()
}

// Delete removes key (best-effort).
func (c *ValkeyCache) Delete(ctx context.Context, key string) {
	_ = c.client.Do(ctx, c.client.B().Del().Key(key).Build()).Error()
}

// Incr atomically increments the integer stored at key (best-effort).
func (c *ValkeyCache) Incr(ctx context.Context, key string) {
	_ = c.client.Do(ctx, c.client.B().Incr().Key(key).Build()).Error()
}
