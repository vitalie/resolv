package resolv

import (
	"net"
	"strings"

	"github.com/miekg/dns"
)

type Request struct {
	Addr  string
	Mode  string
	Name  string
	Type  uint16
	Class uint16
}

type RequestOption func(*Request)

func SetTCPMode(req *Request) {
	req.Mode = "tcp"
}

func SetCHAOSClass(req *Request) {
	req.Class = dns.ClassCHAOS
}

func NewRequest(addr, name string, type_ uint16, options ...RequestOption) *Request {
	if !strings.Contains(addr, ":") {
		addr = net.JoinHostPort(addr, PortDefault)
	}

	req := &Request{
		Addr:  addr,
		Name:  dns.Fqdn(name),
		Type:  type_,
		Class: dns.ClassINET,
	}

	// Apply configuration functions
	for _, opt := range options {
		opt(req)
	}

	return req
}
