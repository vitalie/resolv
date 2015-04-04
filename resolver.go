package resolv

import (
	"net"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Do(ctx context.Context, req *Request) *Response {
	m := new(dns.Msg)
	m.SetQuestion(req.Name, req.Type)

	c := new(dns.Client)
	// TODO: https://godoc.org/github.com/miekg/dns#Exchange
	// - UDP/TCP
	// - Truncated response
	// - Timeout
	// - Retry
	in, rtt, err := c.Exchange(m, req.Addr)
	if err != nil {
		return &Response{Err: err}
	}

	if in.Rcode != dns.RcodeSuccess {
		err := &net.DNSError{Err: dns.RcodeToString[in.Rcode], Name: req.Name, Server: req.Addr}
		return &Response{Err: err}
	}

	return &Response{
		Msg: in,
		Rtt: rtt,
	}
}
