package dnsapi

import (
	"errors"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/webspider/dnsapi/api"
	"github.com/aarpy/wisehoot/crawler/webspider/dnsapi/cache"
	"github.com/aarpy/wisehoot/crawler/webspider/dnsapi/resolve"
)

// NewDNSCache is a great class
func NewDNSCache(dnsServer string, dnsConcurency int, dnsRetryTime string, groupCacheSize int64, redisHost string) api.DNSCache {
	log.Info("Creating new DNS Cache Resolver")

	resolveMgr := resolve.NewResolver(dnsServer, dnsConcurency, 120, dnsRetryTime, true, false)

	log.Info("Creating new DNS Cache Manager")

	cacheMgr := cache.NewCache(groupCacheSize, redisHost, func(request *api.ValueRequest) {

		log.WithField("domain", request.Key).Info("DnsApi:ResolverRequest")

		resolverRequest := api.NewValueRequest(request.Key)
		resolveMgr.Resolve(resolverRequest)
		resolverResponse := <-resolverRequest.Response

		log.WithFields(log.Fields{
			"domain": request.Key,
			"IP":     resolverResponse.Value,
		}).Info("DnsApi:ResolverRequest:Done")

		// Request from DNS
		request.Response <- resolverResponse
	})

	return &dnsCacheMgr{cache: cacheMgr, resolver: resolveMgr}
}

type dnsCacheMgr struct {
	cache    cache.Cache
	resolver resolve.Resolver
}

func (d *dnsCacheMgr) GetIP(domainName string, getIPFunc api.GetIPFunc) {

	log.WithField("domain", domainName).Info("DnsApi:GetIP:Start")

	request := api.NewValueRequest(domainName)

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
				for i, ipString := range ipStrings {
					ipNumbers[i] = net.ParseIP(ipString)
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
