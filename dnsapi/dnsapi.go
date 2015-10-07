package dnsapi

import (
	"errors"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/dnsapi/api"
	"github.com/aarpy/wisehoot/crawler/dnsapi/cache"
	"github.com/aarpy/wisehoot/crawler/dnsapi/resolve"
)

// NewDNSCache is a great class
func NewDNSCache(dnsServer string, dnsConcurency int, dnsRetryTime string, groupCacheSize int64, redisHost string) api.DNSCache {
	log.Info("Creating new DNS Cache Resolver")

	resolveMgr := resolve.NewResolver(dnsServer, dnsConcurency, 120, dnsRetryTime, true, false)

	log.Info("Creating new DNS Cache Manager")

	cacheMgr := cache.NewCache(groupCacheSize, redisHost, func(request *api.ValueRequest) {

		response := api.NewValueResponse("52.1.98.187`", nil)

		log.WithField("IP", response.Value).Info("Resolver response:")

		// Request from DNS
		request.Response <- response

		//return resolveMgr.Resolve(domainName).String()
	})

	return &dnsCacheMgr{cache: cacheMgr, resolver: resolveMgr}
}

type dnsCacheMgr struct {
	cache    cache.Cache
	resolver resolve.Resolver
}

func (d *dnsCacheMgr) GetIP(domainName string, getIPFunc api.GetIPFunc) {

	log.WithField("domain", domainName).Info("DnsApi:GetIP:Start")

	request := api.NewValueRequest(domainName, "DnsApi")

	d.cache.GetValue(request)

	select {
	case response, ok := <-request.Response:
		if ok {
			log.WithFields(log.Fields{
				"domain": domainName,
				"value":  response.Value,
				"err":    response.Err}).Info("DnsApi:GetIP:Response")

			ipStrings := strings.Split(response.Value, " ")
			ipNumbers := make([]net.IP, len(ipStrings))
			if response.Value != "" {
				for _, ipString := range ipStrings {
					ipNumbers = append(ipNumbers, net.ParseIP(ipString))
				}
			}
			getIPFunc(ipNumbers, nil)
		} else {
			log.WithField("domain", domainName).Info("DnsApi:GetIP:ChannelClosed")

			// channel was closed
			getIPFunc(nil, errors.New("Channel closed"))
		}
		break
	case <-time.After(time.Duration(10 * time.Second)):
		log.WithField("domain", domainName).Info("DnsApi:GetIP:ChannelTimeout")
		// channel failed to return in time
		getIPFunc(nil, errors.New("Timeout occurred"))
		break
	}

	log.WithField("domain", domainName).Info("DnsApi:GetIP:Done")
}

func (d *dnsCacheMgr) Invalidate(domainName string) {
	log.WithFields(log.Fields{
		"domainName": domainName,
	}).Info("Invalidate IP")

	d.cache.RemoveValue(domainName)
}
