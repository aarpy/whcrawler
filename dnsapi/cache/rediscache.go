package main

import "gopkg.in/redis.v3"

// NewRedisCache function
func NewRedisCache(hostAddr string, getFunc GetFunc) Cache {
	return &redisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     hostAddr,
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		getFunc: getFunc}
}

type redisCache struct {
	client *redis.Client
}

func (c *redisCache) GetValue(key string) string {
	value, err := c.client.Get(key).Result()
	if err == redis.Nil {
		// key does not exist
		value = c.getFunc(key)
		if value != nil {
			c.Set(key, value, 0)
		}
	} else if err != nil {
		// error occurred
		panic(err)
	}
	return value
}

func (c *redisCache) RemoveValue(key string) string {
	c.client.Del(key)
}

func (c *redisCache) Close() {
	if c.client != nil {
		c.client.Close()
	}
	c.client = nil
}
