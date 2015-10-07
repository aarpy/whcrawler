package cache

// GetFunc to get the key value pairs
type GetFunc func(key string) string

// Cache inteface
type Cache interface {
	GetValue(key string) string
	RemoveValue(key string)
	Close()
}

// NewCache function
func NewCache(cacheSize int64, hostAddr string, getFunc GetFunc) Cache {
	redisCacheMgr := NewRedisCache(hostAddr, getFunc)
	groupCacheMgr := NewGroupCache(cacheSize, func(key string) string {
		return redisCacheMgr.GetValue(key)
	})

	return &cacheMgr{groupCache: groupCacheMgr, redisCache: redisCacheMgr}
}

type cacheMgr struct {
	groupCache Cache
	redisCache Cache
}

func (c *cacheMgr) GetValue(key string) string {
	return c.groupCache.GetValue(key)
}

func (c *cacheMgr) RemoveValue(key string) {
	c.groupCache.RemoveValue(key)
	c.redisCache.RemoveValue(key)
}

func (c *cacheMgr) Close() {
	c.groupCache.Close()
	c.redisCache.Close()
}
