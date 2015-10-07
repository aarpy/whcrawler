package api

import "net"

// GetIPFunc return function
type GetIPFunc func(ips []net.IP, err error)

// DNSCache interface to communicate with the client
type DNSCache interface {
	GetIP(domainName string, getIPFunc GetIPFunc)
	Invalidate(domainName string)
}

// ValueRequest used by cache and resolver
type ValueRequest struct {
	Key      string
	Owner    string
	Response chan *ValueResponse
}

// ValueResponse used by cache and resolver
type ValueResponse struct {
	Value string
	Err   error
}

// NewValueRequest function
func NewValueRequest(key string, owner string) *ValueRequest {
	return &ValueRequest{key, owner, make(chan *ValueResponse, 1)}
}

// NewValueResponse function
func NewValueResponse(value string, err error) *ValueResponse {
	return &ValueResponse{value, err}
}
