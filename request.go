package resolv

import (
	"net"

	"github.com/miekg/dns"
)

type Request struct {
	Addr  string
	Name  string
	Type  uint16
	Class uint16
}

func NewRequest(addr, name string, type_, class uint16) (*Request, error) {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, &net.DNSConfigError{Err: err}
	}

	req := &Request{
		Addr:  addr,
		Name:  dns.Fqdn(name),
		Type:  type_,
		Class: class,
	}

	return req, nil
}
