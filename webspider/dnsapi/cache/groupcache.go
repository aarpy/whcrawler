package cache

import (
	log "github.com/Sirupsen/logrus"
	gc "github.com/golang/groupcache"

	"github.com/aarpy/wisehoot/crawler/webspider/dnsapi/api"
)

const (
	singleGroupName = "shared"
)

type groupCacheMgr struct {
	group gc.Getter
	ctx   gc.Context
}

// NewGroupCache function
func NewGroupCache(cacheSize int64, redisGetFunc GetFunc) Cache {
	return &groupCacheMgr{
		group: gc.NewGroup(singleGroupName, cacheSize, gc.GetterFunc(func(ctx gc.Context, key string, dest gc.Sink) error {
			log.WithField("domain", key).Info("GroupCache:RedisRequest")

			redisRequest := api.NewValueRequest(key)
			redisGetFunc(redisRequest)
			redisResponse := <-redisRequest.Response

			log.WithField("domain", key).Info("GroupCache:RedisRequest:Done")
			return dest.SetString(redisResponse.Value)
		}))}
}

func (c *groupCacheMgr) GetValue(request *api.ValueRequest) {
	log.WithField("domain", request.Key).Info("GroupCache:GetValue")

	var value string
	err := c.group.Get(request, request.Key, gc.StringSink(&value))

	log.WithFields(log.Fields{
		"domain": request.Key,
		"value":  value,
	}).Info("GroupCache:GetValue:Get")

	// send response to client
	request.Response <- api.NewValueResponse(value, err)
	close(request.Response)

	log.WithField("domain", request.Key).Info("GroupCache:GetValue:Done")
}

func (c *groupCacheMgr) RemoveValue(key string) {
	// not available
}

func (c *groupCacheMgr) Close() {
	// do nothing
}
