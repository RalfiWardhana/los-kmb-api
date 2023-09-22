package repository

import (
	"los-kmb-api/domain/cache/interfaces"

	"github.com/allegro/bigcache/v3"
)

type repoHandler struct {
	cache *bigcache.BigCache
}

func NewRepository(cache *bigcache.BigCache) interfaces.Repository {
	return &repoHandler{
		cache: cache,
	}
}

func (c *repoHandler) Get(key string) ([]byte, error) {
	data, err := c.cache.Get(key)
	return data, err
}

func (c *repoHandler) Set(key string, entry []byte) error {
	return c.cache.Set(key, entry)
}
