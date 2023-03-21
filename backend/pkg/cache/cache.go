package cache

import (
	"context"
	"time"

	"github.com/ONLYOFFICE/onlyoffice-pipedrive/pkg/config"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/eko/gocache/lib/v4/store"
	"go-micro.dev/v4/cache"
)

type CustomCache struct {
	store *marshaler.Marshaler
	name  string
}

func (c *CustomCache) Get(ctx context.Context, key string) (interface{}, time.Time, error) {
	var result interface{}
	_, err := c.store.Get(ctx, key, &result)
	return result, time.Now(), err
}

func (c *CustomCache) Put(ctx context.Context, key string, val interface{}, d time.Duration) error {
	return c.store.Set(ctx, key, val, store.WithExpiration(d))
}

func (c *CustomCache) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, key)
}

func (c *CustomCache) String() string {
	return c.name
}

func NewCache(config *config.CacheConfig) cache.Cache {
	switch config.Cache.Type {
	case 1:
		return &CustomCache{
			store: newMemory(config.Cache.Size),
			name:  "Freecache",
		}
	case 2:
		return &CustomCache{
			store: newRedis(config.Cache.Address, config.Cache.Password, 0),
			name:  "Redis",
		}
	default:
		return &CustomCache{
			store: newMemory(config.Cache.Size),
			name:  "Freecache",
		}
	}
}
