package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

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
	log.Info("Wisehoot cralwer started1")

	dnsCache := dnsapi.NewDNSCache(dnsServer, dnsConcurency, dnsRetryTime, groupCacheSize, redisHost)

	dnsCache.GetIP("wisehoot.co", func(ips []net.IP, err error) {
		log.Info("wisehoot.co:", ips, err)
	})
	dnsCache.GetIP("yahoo.com", func(ips []net.IP, err error) {
		log.Info("yahoo.com:", ips, err)
	})
	dnsCache.GetIP("google.com", func(ips []net.IP, err error) {
		log.Info("google.com:", ips, err)
	})
	dnsCache.GetIP("cnn.com", func(ips []net.IP, err error) {
		log.Info("cnn.com:", ips, err)
	})

	time.Sleep(10 * time.Millisecond)
	log.Info("Wisehoot cralwer completed")
}
