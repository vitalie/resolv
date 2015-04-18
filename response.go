package resolv

import (
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type Response struct {
	Addr  string
	Name  string
	Type  uint16
	Class uint16
	Msg   *dns.Msg
	Rtt   time.Duration
	Err   error
}

func NewResponse(req *Request) *Response {
	r := &Response{
		Addr:  req.Addr,
		Name:  req.Name,
		Type:  req.Type,
		Class: req.Class,
	}

	return r
}

func (r *Response) Host() string {
	host, _, err := net.SplitHostPort(r.Addr)
	if err != nil {
		return r.Addr
	}

	return dns.Fqdn(strings.ToLower(host))
}
