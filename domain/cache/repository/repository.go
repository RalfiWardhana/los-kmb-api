package repository

import (
	"errors"
	"los-kmb-api/domain/cache/interfaces"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
)

type repoHandler struct {
	cache             *bigcache.BigCache
	customCaches      map[time.Duration]*bigcache.BigCache
	customCachesMutex sync.RWMutex
}

func NewRepository(cache *bigcache.BigCache) interfaces.Repository {
	return &repoHandler{
		cache:        cache,
		customCaches: make(map[time.Duration]*bigcache.BigCache),
	}
}

func (c *repoHandler) Get(key string) ([]byte, error) {
	data, err := c.cache.Get(key)
	return data, err
}

func (c *repoHandler) Set(key string, entry []byte) error {
	return c.cache.Set(key, entry)
}

func (c *repoHandler) GetWithExpiration(key string) ([]byte, error) {
	// Try the default cache first
	if c.cache != nil {
		data, err := c.cache.Get(key)
		if err == nil {
			return data, nil
		}
	}

	// If not found, try all custom caches
	c.customCachesMutex.RLock()
	defer c.customCachesMutex.RUnlock()

	for _, customCache := range c.customCaches {
		data, err := customCache.Get(key)
		if err == nil {
			return data, nil
		}
	}

	// Not found in any cache
	return nil, errors.New("entry not found")
}

func (c *repoHandler) SetWithExpiration(key string, entry []byte, ttl time.Duration) error {
	// fmt.Printf("DEBUG SetWithExpiration: Called with key=%s, data length=%d, ttl=%v\n", key, len(entry), ttl)

	if c.cache == nil {
		// fmt.Printf("DEBUG SetWithExpiration: ERROR - default cache is nil\n")
		return errors.New("cache is not initialized")
	}

	customCache, err := c.getOrCreateCacheWithTTL(ttl)
	if err != nil {
		// fmt.Printf("DEBUG SetWithExpiration: Failed to get/create cache with TTL %v: %v\n", ttl, err)
		return err
	}

	// fmt.Printf("DEBUG SetWithExpiration: Got custom cache for TTL=%v, storing key=%s\n", ttl, key)

	err = customCache.Set(key, entry)

	// if err != nil {
	// 	fmt.Printf("DEBUG SetWithExpiration: Failed to set key in cache: %v\n", err)
	// } else {
	// 	fmt.Printf("DEBUG SetWithExpiration: Successfully stored key=%s in cache with TTL=%v\n", key, ttl)
	// }

	return err
}

func (c *repoHandler) getOrCreateCacheWithTTL(ttl time.Duration) (*bigcache.BigCache, error) {
	// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Called with TTL=%v\n", ttl)

	if c.cache == nil {
		// fmt.Printf("DEBUG getOrCreateCacheWithTTL: ERROR - default cache is nil\n")
		return nil, errors.New("cache is not initialized")
	}

	c.customCachesMutex.RLock()
	cache, exists := c.customCaches[ttl]
	c.customCachesMutex.RUnlock()

	if exists {
		// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Found existing cache for TTL=%v\n", ttl)
		return cache, nil
	}

	// fmt.Printf("DEBUG getOrCreateCacheWithTTL: No existing cache for TTL=%v, creating new cache\n", ttl)

	c.customCachesMutex.Lock()
	defer c.customCachesMutex.Unlock()

	// Double-check to avoid race conditions
	if cache, exists := c.customCaches[ttl]; exists {
		// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Cache for TTL=%v was created by another goroutine\n", ttl)
		return cache, nil
	}

	// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Creating new BigCache with TTL=%v\n", ttl)
	config := bigcache.DefaultConfig(ttl)
	newCache, err := bigcache.NewBigCache(config)
	if err != nil {
		// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Failed to create BigCache: %v\n", err)
		return nil, err
	}

	// fmt.Printf("DEBUG getOrCreateCacheWithTTL: Successfully created new BigCache with TTL=%v\n", ttl)
	c.customCaches[ttl] = newCache
	return newCache, nil
}
