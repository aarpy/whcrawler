package main

import (
	"fmt"
	"os"
	"strconv"

	gc "github.com/golang/groupcache"
)

const (
	stringGroupName = "TestGroup"
	cacheSize       = 1 << 20
)

var (
	stringGroup gc.Getter
	dummyCtx    gc.Context
)

type Cache interface {
	Init(getterFunc func)
	GetIP(domainName) net.IP
}

type cacheMgr struct {
	getterFunc	func
	stringGroup gc.Getter
	dummyCtx    gc.Context
}

func NewCache() Cache  {
	return &cacheMgr{}
}

func main() {

	// setup group
	stringGroup = gc.NewGroup(stringGroupName, cacheSize, gc.GetterFunc(func(_ gc.Context, key string, dest gc.Sink) error {
		fmt.Fprintf(os.Stdout, "Setting cache key: %s\n", key)
		return dest.SetString("ECHO:" + key)
	}))

	// Get Items
	for j := 0; j < 4; j++ {
		for i := 0; i < 4; i++ {
			var s string
			if err := stringGroup.Get(dummyCtx, "TestCaching-key"+strconv.Itoa(i), gc.StringSink(&s)); err != nil {
				fmt.Fprintf(os.Stdout, "TestCaching-key value: failed%s\n", err)
				return
			}
			fmt.Fprintf(os.Stdout, "TestCaching-key value:%s\n", s)
		}
		fmt.Fprintln(os.Stdout, "---")
	}

	fmt.Fprintln(os.Stdout, "Done.")
}
