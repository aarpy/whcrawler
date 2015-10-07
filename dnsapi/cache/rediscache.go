package cache

import (
	log "github.com/Sirupsen/logrus"
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
	log.WithField("domain", request.Key).Info("RedisCache:GetValue:Start")

	// Get from Redis
	value, err := c.client.Get(request.Key).Result()

	log.WithFields(log.Fields{
		"domain": request.Key,
		"value":  value,
		"err":    err,
		"owner":  request.Owner,
	}).Info("RedisCache:GetValue:GetComplete")

	// Value found without error or empty
	if err == nil {
		// Notify the calling group cache
		request.Response <- api.NewValueResponse(value, nil)
		return
	}

	log.WithField("domain", request.Key).Info("RedisCache:GetValue:CheckResolver")

	// Request Resolver
	resolverRequest := api.NewValueRequest(request.Key, "RedisCache")
	c.resolverFunc(resolverRequest)
	resolverResponse := <-resolverRequest.Response

	log.WithFields(log.Fields{
		"domain": request.Key,
		"value":  value,
		"owner":  request.Owner,
	}).Info("RedisCache:GetValue:FromResolver")

	// Save it to Redis irrespectively to ensure no requests are sent to Resolver
	c.client.Set(request.Key, resolverResponse.Value, 0)

	log.WithField("domain", request.Key).Info("RedisCache:GetValue:SetComplete")

	// Notify the calling group cache
	request.Response <- resolverResponse

	log.WithField("domain", request.Key).Info("RedisCache:GetValue:Done")
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
