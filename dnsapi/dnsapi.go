package dnsapi

import (
	"fmt"
	"net"
	"os"

	"dnsapi/resolve"
)

// DNSCache interface to communicate with the client
type DNSCache interface {
	GetIP(domain string) net.IP
	InvalidateIP(domain string, ip net.IP) net.IP
}

// NewDNSCache is a great class
func NewDNSCache() DNSCache {
	fmt.Fprintln(os.Stdout, "Wisehoot cralwer started")
	return nil
}

type dnsCacheMgr struct {
	resolve.Resolver
}

func (d *dnsCacheMgr) GetIP(domain string) net.IP {
	return nil
}

func (d *dnsCacheMgr) InvalidateIP(domain string, ip net.IP) net.IP {
	return nil
}
