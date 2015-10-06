package main

import gc "github.com/golang/groupcache"

const (
	singleGroupName = "shared"
)

type groupCacheMgr struct {
	group gc.Getter
	ctx   gc.Context
}

// NewGroupCache function
func NewGroupCache(cacheSize int, getFunc GetFunc) Cache {
	return &groupCacheMgr{
		group: gc.NewGroup(singleGroupName, cacheSize, gc.GetterFunc(func(_ gc.Context, key string, dest gc.Sink) error {
			return dest.SetString(getFunc(key))
		}))}
}

func (c *groupCacheMgr) GetValue(key string) string {
	var value string
	if err := stringGroup.Get(ctx, key, gc.StringSink(&value)); err != nil {
		return nil
	}
	return value
}

func (c *groupCacheMgr) RemoveValue(key string) string {
	c.stringGroup.RemoveValue(key)
}

func (c *groupCacheMgr) Close() {
	// do nothing
}
