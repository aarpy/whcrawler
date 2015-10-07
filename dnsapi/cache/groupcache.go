package cache

import gc "github.com/golang/groupcache"

const (
	singleGroupName = "shared"
)

type groupCacheMgr struct {
	group gc.Getter
	ctx   gc.Context
}

// NewGroupCache function
func NewGroupCache(cacheSize int64, getFunc GetFunc) Cache {
	return &groupCacheMgr{
		group: gc.NewGroup(singleGroupName, cacheSize, gc.GetterFunc(func(_ gc.Context, key string, dest gc.Sink) error {
			return dest.SetString(getFunc(key))
		}))}
}

func (c *groupCacheMgr) GetValue(key string) string {
	var value string
	if err := c.group.Get(c.ctx, key, gc.StringSink(&value)); err != nil {
		return ""
	}
	return value
}

func (c *groupCacheMgr) RemoveValue(key string) {
	// not available
}

func (c *groupCacheMgr) Close() {
	// do nothing
}
