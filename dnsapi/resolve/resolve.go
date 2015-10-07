package resolve

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/aarpy/wisehoot/crawler/dnsapi/api"
)

func doReadDomains(request *api.ValueRequest, requests chan<- *api.ValueRequest, domainSlotAvailable <-chan bool) {
	for _ = range domainSlotAvailable {

		request.Key += "."

		requests <- request

		break
	}
}

var sendingDelay time.Duration
var retryDelay time.Duration

// Resolver interface
type Resolver interface {
	Resolve(request *api.ValueRequest)
}

type resolveMgr struct {
	dnsServer           string
	concurrency         int
	packetsPerSecond    int
	retryTime           string
	verbose             bool
	ipv6                bool
	conn                net.Conn
	requests            chan *api.ValueRequest
	domainSlotAvailable chan bool
}

// NewResolver function
func NewResolver(dnsServer string, concurrency int, packetsPerSecond int, retryTime string, verbose bool, ipv6 bool) Resolver {
	mgr := &resolveMgr{dnsServer, concurrency, packetsPerSecond, retryTime, verbose, ipv6, nil, nil, nil}

	go mgr.init()

	return mgr
}

func (r *resolveMgr) init() {

	sendingDelay = time.Duration(1000000000/r.packetsPerSecond) * time.Nanosecond
	var err error
	retryDelay, err = time.ParseDuration(r.retryTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse duration %s\n", r.retryTime)
		panic(err)
	}

	fmt.Fprintf(os.Stderr, "Server: %s, sending delay: %s (%d pps), retry delay: %s\n",
		r.dnsServer, sendingDelay, r.packetsPerSecond, retryDelay)

	r.requests = make(chan *api.ValueRequest, r.concurrency)
	r.domainSlotAvailable = make(chan bool, r.concurrency)

	for i := 0; i < r.concurrency; i++ {
		r.domainSlotAvailable <- true
	}

	c, err := net.Dial("udp", r.dnsServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bind(udp, %s): %s\n", r.dnsServer, err)
		panic(err)
	}

	// Used as a queue. Make sure it has plenty of storage available.
	timeoutRegister := make(chan *domainRecord, r.concurrency*1000)
	timeoutExpired := make(chan *domainRecord)

	resolved := make(chan *domainAnswer, r.concurrency)
	tryResolving := make(chan *domainRecord, r.concurrency)

	log.Info("Running resolver functions")
	go doTimeouter(timeoutRegister, timeoutExpired)

	go doSend(c, tryResolving, r.ipv6)
	go doReceive(c, resolved, r.ipv6)

	log.Info("Running Map functions")
	t0 := time.Now()
	domainsCount, avgTries := doMapGaurd(r.requests, r.domainSlotAvailable,
		timeoutRegister, timeoutExpired,
		tryResolving, resolved, r.verbose)
	td := time.Now().Sub(t0)
	fmt.Fprintf(os.Stderr, "Resolved %d domains in %.3fs. Average retries %.3f. Domains per second: %.3f\n",
		domainsCount,
		td.Seconds(),
		avgTries,
		float64(domainsCount)/td.Seconds())

	log.Info("Resolver: Initilized")
}

// Resolve function to resolve a domain name into IP address
func (r *resolveMgr) Resolve(request *api.ValueRequest) {

	log.WithField("domain", request.Key).Info("Resolver:Resolve")

	doReadDomains(request, r.requests, r.domainSlotAvailable)

}

type domainRecord struct {
	id      uint16
	request *api.ValueRequest
	timeout time.Time
	resend  int
}

type domainAnswer struct {
	id     uint16
	domain string
	ips    []net.IP
}

func doMapGaurd(requests <-chan *api.ValueRequest,
	domainSlotAvailable chan<- bool,
	timeoutRegister chan<- *domainRecord,
	timeoutExpired <-chan *domainRecord,
	tryResolving chan<- *domainRecord,
	resolved <-chan *domainAnswer,
	verbose bool) (int, float64) {

	m := make(map[uint16]*domainRecord)

	done := false

	sumTries := 0
	domainCount := 0

	for done == false || len(m) > 0 {
		select {
		case domain := <-requests:
			fmt.Fprintf(os.Stdout, "Found domain: %s\n", domain)
			var id uint16
			for {
				id = uint16(rand.Int())
				if id != 0 && m[id] == nil {
					break
				}
			}
			dr := &domainRecord{id, domain, time.Now(), 1}
			m[id] = dr
			if verbose {
				fmt.Fprintf(os.Stderr, "0x%04x resolving %s\n", id, domain)
			}
			timeoutRegister <- dr
			tryResolving <- dr

		case dr := <-timeoutExpired:
			if m[dr.id] == dr {
				dr.resend++
				dr.timeout = time.Now()
				if verbose {
					fmt.Fprintf(os.Stderr, "0x%04x resend (try:%d) %s\n", dr.id,
						dr.resend, dr.request.Key)
				}
				timeoutRegister <- dr
				tryResolving <- dr
			}

		case da := <-resolved:
			if m[da.id] != nil {
				dr := m[da.id]
				if dr.request.Key != da.domain {
					if verbose {
						fmt.Fprintf(os.Stderr, "0x%04x error, unrecognized domain: %s != %s\n",
							da.id, dr.request.Key, da.domain)
					}
					break
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "0x%04x resolved %s\n",
						dr.id, dr.request.Key)
				}

				s := make([]string, 0, 16)
				for _, ip := range da.ips {
					s = append(s, ip.String())
				}
				sort.Sort(sort.StringSlice(s))

				// without trailing dot
				dr.request.Key = dr.request.Key[:len(dr.request.Key)-1]
				ips := strings.Join(s, " ")
				fmt.Printf("%s, %s\n", dr.request.Key, ips)

				dr.request.Response <- api.NewValueResponse(ips, nil)

				sumTries += dr.resend
				domainCount++

				delete(m, dr.id)
				domainSlotAvailable <- true
			}
		}
	}
	return domainCount, float64(sumTries) / float64(domainCount)
}

func doTimeouter(timeoutRegister <-chan *domainRecord,
	timeoutExpired chan<- *domainRecord) {
	for {
		dr := <-timeoutRegister
		t := dr.timeout.Add(retryDelay)
		now := time.Now()
		if t.Sub(now) > 0 {
			delta := t.Sub(now)
			time.Sleep(delta)
		}
		timeoutExpired <- dr
	}
}

func doSend(c net.Conn, tryResolving <-chan *domainRecord, ipv6 bool) {
	for {
		dr := <-tryResolving

		var t uint16
		if !ipv6 {
			t = dnsTypeA
		} else {
			t = dnsTypeAAAA
		}
		msg := packDns(dr.request.Key, dr.id, t)

		_, err := c.Write(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write(udp): %s\n", err)
			os.Exit(1)
		}
		time.Sleep(sendingDelay)
	}
}

func doReceive(c net.Conn, resolved chan<- *domainAnswer, ipv6 bool) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		var t uint16
		if !ipv6 {
			t = dnsTypeA
		} else {
			t = dnsTypeAAAA
		}
		domain, id, ips := unpackDns(buf[:n], t)
		resolved <- &domainAnswer{id, domain, ips}
	}
}
