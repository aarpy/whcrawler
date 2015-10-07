package dnsapi

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/aarpy/wisehoot/crawler/dnsapi/cache"
	"github.com/aarpy/wisehoot/crawler/dnsapi/resolve"
)

// DNSCache interface to communicate with the client
type DNSCache interface {
	GetIP(domainName string) []net.IP
	Invalidate(domainName string)
}

// NewDNSCache is a great class
func NewDNSCache(dnsServer string, dnsConcurency int, dnsRetryTime string, groupCacheSize int64, redisHost string) DNSCache {
	fmt.Fprintln(os.Stdout, "Wisehoot cralwer started")

	resolveMgr := resolve.NewResolver(dnsServer, dnsConcurency, 120, dnsRetryTime, true, false)
	cacheMgr := cache.NewCache(groupCacheSize, redisHost, func(domainName string) string {
		return resolveMgr.Resolve(domainName).String()
	})

	return &dnsCacheMgr{cache: cacheMgr, resolver: resolveMgr}
}

type dnsCacheMgr struct {
	cache    cache.Cache
	resolver resolve.Resolver
}

func (d *dnsCacheMgr) GetIP(domainName string) []net.IP {
	var ipNumbers []net.IP

	if ipStrings := d.cache.GetValue(domainName); ipStrings != "" {
		for _, ipString := range strings.Split(ipStrings, " ") {
			ipNumbers = append(ipNumbers, net.ParseIP(ipString))
		}
		return ipNumbers
	}
	return nil
}

func (d *dnsCacheMgr) Invalidate(domainName string) {
	d.cache.RemoveValue(domainName)
}
