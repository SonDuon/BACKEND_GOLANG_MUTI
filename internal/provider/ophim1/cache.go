package ophim1

import (
	"sync"
	"time"

	"github.com/SonDuon/BACKEND_GOLANG_MUTI/internal/provider"
)

type metaCache struct {
	mu   sync.RWMutex
	data map[string]*cachedMeta
	ttl  time.Duration
}
type cachedMeta struct {
	dto *provider.MovieDTO
	exp time.Time
}

func newMetaCache(ttl time.Duration) *metaCache {
	return &metaCache{data: make(map[string]*cachedMeta), ttl: ttl}
}
func (c *metaCache) get(key string) *provider.MovieDTO {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.data[key]; ok && time.Now().Before(v.exp) {
		return v.dto
	}
	return nil
}
func (c *metaCache) set(key string, dto *provider.MovieDTO) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = &cachedMeta{dto: dto, exp: time.Now().Add(c.ttl)}
}
func (c *metaCache) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}
