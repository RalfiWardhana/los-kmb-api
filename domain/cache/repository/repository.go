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

	if c.cache == nil {
		return errors.New("cache is not initialized")
	}

	customCache, err := c.getOrCreateCacheWithTTL(ttl)
	if err != nil {
		return err
	}

	err = customCache.Set(key, entry)

	return err
}

func (c *repoHandler) getOrCreateCacheWithTTL(ttl time.Duration) (*bigcache.BigCache, error) {

	if c.cache == nil {
		return nil, errors.New("cache is not initialized")
	}

	c.customCachesMutex.RLock()
	cache, exists := c.customCaches[ttl]
	c.customCachesMutex.RUnlock()

	if exists {
		return cache, nil
	}

	c.customCachesMutex.Lock()
	defer c.customCachesMutex.Unlock()

	// Double-check to avoid race conditions
	if cache, exists := c.customCaches[ttl]; exists {
		return cache, nil
	}

	config := bigcache.DefaultConfig(ttl)
	newCache, err := bigcache.NewBigCache(config)
	if err != nil {
		return nil, err
	}

	c.customCaches[ttl] = newCache
	return newCache, nil
}
