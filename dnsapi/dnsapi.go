package dnsapi

import (
	"net"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/dnsapi/cache"
	"github.com/aarpy/wisehoot/crawler/dnsapi/resolve"
)

// GetIPFunc return function
type GetIPFunc func(ips []net.IP, err error)

// DNSCache interface to communicate with the client
type DNSCache interface {
	GetIP(domainName string, getIPFunc GetIPFunc)
	Invalidate(domainName string)
}

// NewDNSCache is a great class
func NewDNSCache(dnsServer string, dnsConcurency int, dnsRetryTime string, groupCacheSize int64, redisHost string) DNSCache {
	log.Info("Creating new DNS Cache")

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

func (d *dnsCacheMgr) GetIP(domainName string, getIPFunc GetIPFunc) {
	var ipNumbers []net.IP

	log.WithFields(log.Fields{
		"domainName": domainName,
	}).Info("Get IP")

	if ipStrings := d.cache.GetValue(domainName); ipStrings != "" {
		for _, ipString := range strings.Split(ipStrings, " ") {
			ipNumbers = append(ipNumbers, net.ParseIP(ipString))
		}
	}
	getIPFunc(ipNumbers, nil)
}

func (d *dnsCacheMgr) Invalidate(domainName string) {
	log.WithFields(log.Fields{
		"domainName": domainName,
	}).Info("Invalidate IP")

	d.cache.RemoveValue(domainName)
}
