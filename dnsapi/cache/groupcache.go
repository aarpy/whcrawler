package cache

import (
	gc "github.com/golang/groupcache"

	"github.com/aarpy/wisehoot/crawler/dnsapi/api"
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

			redisRequest := api.NewValueRequest(key)
			redisGetFunc(redisRequest)
			redisResponse := <-redisRequest.Response

			return dest.SetString(redisResponse.Value)
		}))}
}

func (c *groupCacheMgr) GetValue(request *api.ValueRequest) {

	var value string
	err := c.group.Get(request, request.Key, gc.StringSink(&value))

	// send response to client
	request.Response <- api.NewValueResponse(value, err)
	close(request.Response)
}

func (c *groupCacheMgr) RemoveValue(key string) {
	// not available
}

func (c *groupCacheMgr) Close() {
	// do nothing
}
