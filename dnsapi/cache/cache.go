package cache

import (
	"fmt"
)

// GetFunc to get the key value pairs
type GetFunc func(key string) string

// Cache inteface
type Cache interface {
	GetValue(key string) string
	RemoveValue(key string)
	Close()
}

// NewCache function
func NewCache(cacheSize int, hostAddr string, getFunc GetFunc) Cache {
	fmt.Println("Creating new cache")
	return &cacheMgr{
		groupCache: NewGroupCache(cacheSize, func(key string) string {
			return c.redisCache.GetValue(key)
		}),
		redisCache: NewRedisCache(hostAddr, getFunc)}
}

type cacheMgr struct {
	groupCache Cache
	redisCache Cache
}

func (c *cacheMgr) GetValue(key string) string {
	return c.groupCache.GetValue(key)
}

func (c *cacheMgr) RemoveValue(key string) string {
	c.groupCache.RemoveValue(key)
	c.redisCache.RemoveValue(key)
}

func (c *cacheMgr) Close() {
	c.groupCache.Close()
	c.redisCache.Close()
}
