package cache

import (
	"gopkg.in/redis.v3"

	"github.com/aarpy/wisehoot/crawler/dnsapi/api"
)

// NewRedisCache function
func NewRedisCache(hostAddress string, getFunc GetFunc) Cache {
	return &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     hostAddress,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		resolverFunc: getFunc}
}

type redisCache struct {
	client       *redis.Client
	resolverFunc GetFunc
}

func (c *redisCache) GetValue(request *api.ValueRequest) {
	// Get from Redis
	value, err := c.client.Get(request.Key).Result()

	// Value found
	if err == redis.Nil {
		// Notify the calling group cache
		request.Response <- api.NewValueResponse(value, nil)
		return
	}

	// Request Resolver
	resolverRequest := api.NewValueRequest(request.Key)
	c.resolverFunc(resolverRequest)
	resolverResponse := <-resolverRequest.Response

	// Save it to Redis irrespectively to ensure no requests are sent to Resolver
	c.client.Set(request.Key, resolverResponse.Value, 0)

	// Notify the calling group cache
	request.Response <- resolverResponse
}

func (c *redisCache) RemoveValue(key string) {
	c.client.Del(key)
}

func (c *redisCache) Close() {
	if c.client != nil {
		c.client.Close()
	}
	c.client = nil
}
