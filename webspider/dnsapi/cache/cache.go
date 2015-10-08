package cache

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/webspider/dnsapi/api"
)

// GetFunc to get the key value pairs
type GetFunc func(request *api.ValueRequest)

// Cache interface
type Cache interface {
	GetValue(request *api.ValueRequest)
	RemoveValue(key string)
	Close()
}

// NewCache function
func NewCache(cacheSize int64, hostAddress string, getFunc GetFunc) Cache {
	redisCacheMgr := NewRedisCache(hostAddress, getFunc)

	groupCacheMgr := NewGroupCache(cacheSize, func(request *api.ValueRequest) {
		redisCacheMgr.GetValue(request)
	})

	return &cacheMgr{groupCache: groupCacheMgr, redisCache: redisCacheMgr}
}

type cacheMgr struct {
	groupCache Cache
	redisCache Cache
}

func (c *cacheMgr) GetValue(request *api.ValueRequest) {
	log.WithField("domain", request.Key).Info("Cache:GetValue")

	c.groupCache.GetValue(request)
}

func (c *cacheMgr) RemoveValue(key string) {
	c.groupCache.RemoveValue(key)
	c.redisCache.RemoveValue(key)
}

func (c *cacheMgr) Close() {
	c.groupCache.Close()
	c.redisCache.Close()
}
