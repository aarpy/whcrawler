package dnsapi

import (
	"fmt"
	"net"
	"os"
)

const (
	groupCacheSize = 1 << 20
)

// DNSCache interface to communicate with the client
type DNSCache interface {
	GetIP(domainName string) []net.IP
	Invalidate(domainName string)
}

// NewDNSCache is a great class
func NewDNSCache() DNSCache {
	fmt.Fprintln(os.Stdout, "Wisehoot cralwer started")

	resolver := NewResolver()

	return &dnsCacheMgr{
		cache: NewCache(groupCacheSize, "localhost:6379", func(domainName string) string {
			return resolver.resolve(domainName)
		}),
		resovler: resolver}
}

type dnsCacheMgr struct {
	cache    Cache
	resolver resolve.Resolver
}

func (d *dnsCacheMgr) GetIP(domainName string) []net.IP {
	var ipNumbers []net.IP

	if ipStrings := d.cache(domainName); ipStrings != nil {
		for _, ipString := range Split(ipStrings, " ") {
			append(ipNumbers, ParseIP(ipString))
		}
		return ipNumbers
	}
	return nil
}

func (d *dnsCacheMgr) Invalidate(domainName string) {
	d.cache.RemoveValue(domainName)
}
