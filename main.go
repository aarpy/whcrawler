package main

import (
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/dnsapi"
)

const (
	dnsServer      = "8.8.8.8:53"
	dnsConcurrency = 5000
	dnsRetryTime   = "1s"
	groupCacheSize = 1 << 20
	redisHost      = "localhost:6379"
)

func main() {
	log.Info("Wisehoot cralwer started1")

	dnsCache := dnsapi.NewDNSCache(dnsServer, dnsConcurrency, dnsRetryTime, groupCacheSize, redisHost)

	domains := []string{"wisehoot.co", "google.com", "microsoft.com", "industrycharlotte.com"}
	for _, domain := range domains {
		dnsCache.GetIP(domain, func(ips []net.IP, err error) {
			log.WithFields(log.Fields{
				"domain":  domain,
				"ips":     ips,
				"ips_len": len(ips),
			}).Info("Main:Domain:Complete")
		})
	}

	log.Info("Wisehoot cralwer completed")
}
