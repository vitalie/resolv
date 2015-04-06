package resolv

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Do(ctx context.Context, req *Request) <-chan *Response {
	c := make(chan *Response, 1)

	go func() {
		defer func() {
			// Unhandled error
			if err := recover(); err != nil {
				c <- &Response{Err: fmt.Errorf("resolve: %v", err)}
			}
			close(c)
		}()

		// Prepare message
		m := new(dns.Msg)
		m.SetQuestion(req.Name, req.Type)

		cli := new(dns.Client)
		// TODO: https://godoc.org/github.com/miekg/dns#Exchange
		// - UDP/TCP
		// - Truncated response
		// - Timeout
		// - Retry
		in, rtt, err := cli.Exchange(m, req.Addr)
		if err != nil {
			c <- &Response{Err: err}
			return
		}

		if in.Rcode != dns.RcodeSuccess {
			err := &net.DNSError{Err: dns.RcodeToString[in.Rcode], Name: req.Name, Server: req.Addr}
			c <- &Response{Err: err}
			return
		}

		c <- &Response{
			Msg: in,
			Rtt: rtt,
		}
	}()

	return c
}
