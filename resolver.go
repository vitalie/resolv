package resolv

import (
	"fmt"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(req *Request) <-chan *Response {
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
		// m := new(dns.Msg)
		// m.SetQuestion(req.Name, req.Type)
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = true
		m.Question = make([]dns.Question, 1)
		m.Question[0] = dns.Question{req.Name, req.Type, req.Class}

		cli := new(dns.Client)
		cli.Net = req.Mode

		// TODO:
		// https://godoc.org/github.com/miekg/dns#Client
		// - Truncated response
		// - Timeout
		// - Retry

		in, rtt, err := cli.Exchange(m, req.Addr)
		if err != nil {
			c <- &Response{Err: err}
			return
		}

		// DNS error
		if in.Rcode != dns.RcodeSuccess {
			err := NewDNSError(
				dns.RcodeToString[in.Rcode],
				req,
			)
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

func (r *Resolver) ResolveTypes(ctx context.Context, proto, addr string, name string, types ...uint16) <-chan *Response {
	c := make(chan *Response, len(types))

	// TODO: implement me
	for i := 0; i < len(types); i++ {
		c <- &Response{}
	}

	close(c)
	return c
}

func (r *Resolver) ResolveNames(ctx context.Context, proto, addr string, type_ uint16, names ...string) <-chan *Response {
	c := make(chan *Response, len(names))

	// TODO: implement me
	for i := 0; i < len(names); i++ {
		c <- &Response{}
	}

	close(c)
	return c
}
