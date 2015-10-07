package resolve

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

func doReadDomains(domains chan<- string, domainSlotAvailable <-chan bool) {
	in := bufio.NewReader(os.Stdin)

	for _ = range domainSlotAvailable {

		input, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "read(stdin): %s\n", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		domain := input + "."

		domains <- domain
	}
	close(domains)
}

var sendingDelay time.Duration
var retryDelay time.Duration

// Resolver interface
type Resolver interface {
	Resolve(domainName string) int
}

type resolveMgr struct {
	dnsServer        string
	concurrency      int
	packetsPerSecond int
	retryTime        string
	verbose          bool
	ipv6             bool
	conn             net.Conn
}

// NewResolver function
func NewResolver(dnsServer string, concurrency int, packetsPerSecond int, retryTime string, verbose bool, ipv6 bool) Resolver {
	return &resolveMgr{dnsServer, concurrency, packetsPerSecond, retryTime, verbose, ipv6, nil}
}

// Resolve function to resolve a domain name into IP address
func (r *resolveMgr) Resolve(domainName string) int {

	sendingDelay = time.Duration(1000000000/r.packetsPerSecond) * time.Nanosecond
	var err error
	retryDelay, err = time.ParseDuration(r.retryTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse duration %s\n", r.retryTime)
		return 0
	}

	fmt.Fprintf(os.Stderr, "Server: %s, sending delay: %s (%d pps), retry delay: %s\n",
		r.dnsServer, sendingDelay, r.packetsPerSecond, retryDelay)

	domains := make(chan string, r.concurrency)
	domainSlotAvailable := make(chan bool, r.concurrency)

	for i := 0; i < r.concurrency; i++ {
		domainSlotAvailable <- true
	}

	go doReadDomains(domains, domainSlotAvailable)

	c, err := net.Dial("udp", r.dnsServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bind(udp, %s): %s\n", r.dnsServer, err)
		os.Exit(1)
	}

	// Used as a queue. Make sure it has plenty of storage available.
	timeoutRegister := make(chan *domainRecord, r.concurrency*1000)
	timeoutExpired := make(chan *domainRecord)

	resolved := make(chan *domainAnswer, r.concurrency)
	tryResolving := make(chan *domainRecord, r.concurrency)

	go doTimeouter(timeoutRegister, timeoutExpired)

	go doSend(c, tryResolving, r.ipv6)
	go doReceive(c, resolved, r.ipv6)

	t0 := time.Now()
	domainsCount, avgTries := doMapGaurd(domains, domainSlotAvailable,
		timeoutRegister, timeoutExpired,
		tryResolving, resolved, r.verbose)
	td := time.Now().Sub(t0)
	fmt.Fprintf(os.Stderr, "Resolved %d domains in %.3fs. Average retries %.3f. Domains per second: %.3f\n",
		domainsCount,
		td.Seconds(),
		avgTries,
		float64(domainsCount)/td.Seconds())

	return 1
}

type domainRecord struct {
	id      uint16
	domain  string
	timeout time.Time
	resend  int
}

type domainAnswer struct {
	id     uint16
	domain string
	ips    []net.IP
}

func doMapGaurd(domains <-chan string,
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
		case domain := <-domains:
			fmt.Fprintf(os.Stdout, "Found domain: %s\n", domain)
			if domain == "" {
				domains = make(chan string)
				done = true
				break
			}
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
						dr.resend, dr.domain)
				}
				timeoutRegister <- dr
				tryResolving <- dr
			}

		case da := <-resolved:
			if m[da.id] != nil {
				dr := m[da.id]
				if dr.domain != da.domain {
					if verbose {
						fmt.Fprintf(os.Stderr, "0x%04x error, unrecognized domain: %s != %s\n",
							da.id, dr.domain, da.domain)
					}
					break
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "0x%04x resolved %s\n",
						dr.id, dr.domain)
				}

				s := make([]string, 0, 16)
				for _, ip := range da.ips {
					s = append(s, ip.String())
				}
				sort.Sort(sort.StringSlice(s))

				// without trailing dot
				domain := dr.domain[:len(dr.domain)-1]
				fmt.Printf("%s, %s\n", domain, strings.Join(s, " "))

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
		msg := packDns(dr.domain, dr.id, t)

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
