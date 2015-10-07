package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aarpy/wisehoot/crawler/dnsapi"
)

const (
	dnsServer      = "8.8.8.8:53"
	dnsConcurency  = 5000
	dnsRetryTime   = "1s"
	groupCacheSize = 1 << 20
	redisHost      = "localhost:6379"
)

func main() {
	fmt.Fprintln(os.Stdout, "Wisehoot cralwer started")

	dnsCache := dnsapi.NewDNSCache(dnsServer, dnsConcurency, dnsRetryTime, groupCacheSize, redisHost)
	dnsCache.GetIP("wisehoot.co")
	dnsCache.GetIP("wisehoot.co")
	dnsCache.GetIP("wisehoot.co")
	dnsCache.GetIP("wisehoot.co")

	thread.Sleep(10 * time.Millisecond)
}
