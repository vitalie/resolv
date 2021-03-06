package resolv

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

// Request represents a DNS request.
type Request struct {
	Addr    string // Remote host:port
	Mode    string // Mode tcp/udp
	Name    string // Query name
	Type    uint16 // Query type
	Class   uint16 // Query class
	Recurse bool   // Recursion desired
}

// RequestOption represents an option function.
type RequestOption func(*Request)

// SetTCPMode configures the query to run in TCP mode.
func SetTCPMode(req *Request) {
	req.Mode = "tcp"
}

// SetCHAOSClass sets the class to CHAOS for a request.
func SetCHAOSClass(req *Request) {
	req.Class = dns.ClassCHAOS
}

// SetRD enables recursion for a request.
func SetRD(req *Request) {
	req.Recurse = true
}

func NewRequest(addr, name string, type_ uint16, options ...RequestOption) *Request {
	if _, _, err := net.SplitHostPort(addr); err != nil {
		addr = net.JoinHostPort(addr, DefaultPort)

	}

	req := &Request{
		Addr:  addr,
		Name:  dns.Fqdn(strings.ToLower(name)),
		Type:  type_,
		Class: dns.ClassINET,
	}

	// Apply configuration functions
	for _, opt := range options {
		opt(req)
	}

	return req
}
