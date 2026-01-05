package cache

import (
	"encoding/json"
	"time"

	"github.com/dgraph-io/ristretto"
)

// Cache wraps Ristretto cache with helper methods
type Cache struct {
	client *ristretto.Cache
}

// Config holds cache configuration
type Config struct {
	MaxCost     int64         // Maximum cache size in bytes (e.g., 100MB)
	NumCounters int64         // Number of counters for frequency tracking (e.g., 10M)
	BufferItems int64         // Ring buffer size for async operations
	DefaultTTL  time.Duration // Default TTL for cache entries
}

// NewCache creates a new cache instance
func NewCache(config Config) (*Cache, error) {
	ristrettoConfig := &ristretto.Config{
		NumCounters: config.NumCounters,
		MaxCost:     config.MaxCost,
		BufferItems: config.BufferItems,
	}

	client, err := ristretto.NewCache(ristrettoConfig)
	if err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
	}, nil
}

// Set stores a value in cache with automatic cost calculation
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	// Calculate cost based on JSON size
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	cost := int64(len(jsonBytes))

	// Set with TTL
	c.client.SetWithTTL(key, value, cost, ttl)

	// Wait for value to pass through buffers
	c.client.Wait()

	return nil
}

// Get retrieves a value from cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.client.Get(key)
}

// Delete removes a value from cache
func (c *Cache) Delete(key string) {
	c.client.Del(key)
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.client.Clear()
}

// Close closes the cache
func (c *Cache) Close() {
	c.client.Close()
}

// GetMetrics returns cache metrics
func (c *Cache) GetMetrics() *ristretto.Metrics {
	return c.client.Metrics
}
