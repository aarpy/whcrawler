package main

import (
	"net"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/agent"
	"github.com/aarpy/wisehoot/crawler/dnsapi"
)

const (
	dnsServer      = "8.8.8.8:53"
	dnsConcurrency = 5000
	dnsRetryTime   = "1s"
	groupCacheSize = 1 << 20
	redisHost      = "localhost:6379"
)

func mainDNS() {
	log.Info("Wisehoot cralwer started1")

	dnsCache := dnsapi.NewDNSCache(dnsServer, dnsConcurrency, dnsRetryTime, groupCacheSize, redisHost)

	domains := []string{"wisehoot.co", "google.com", "microsoft.com", "cnn.com", "industrycharlotte.com"}
	for _, domain := range domains {
		go dnsCache.GetIP(domain, func(ips []net.IP, err error) {
			log.WithFields(log.Fields{
				"domain":  domain,
				"ips":     ips,
				"ips_len": len(ips),
			}).Info("Main:Domain:Complete")
		})
	}

	time.Sleep(5 * time.Second)
	log.Info("Wisehoot cralwer completed")
}

func main() {
	_ = agent.NewPastebin(time.Minute, "c980e3546ef3099f8e9cc68f3cce3a62")
}
