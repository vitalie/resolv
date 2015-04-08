package resolv

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type Resolver struct {
}

func NewResolver() *Resolver {
	return &Resolver{}
}

func (r *Resolver) Resolve(req *Request) <-chan *Response {
	// Buffered channel to avoid goroutine leaking
	c := make(chan *Response, 1)

	go func() {
		defer func() {
			// Unhandled error
			if err := recover(); err != nil {
				c <- NewResponseErr(req, fmt.Errorf("resolve: %v", err))
			}
			close(c)
		}()

		// Prepare message
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
			c <- NewResponseErr(req, err)
			return
		}

		// DNS error
		if in.Rcode != dns.RcodeSuccess {
			err := NewDNSError(
				dns.RcodeToString[in.Rcode],
				req,
			)

			if in.Rcode == dns.RcodeNameError {
				err.IsNameError = true
			}

			c <- NewResponseErr(req, err)
			return
		}

		// Truncation
		if in.MsgHdr.Truncated {
			err := NewDNSError(
				"truncated",
				req,
			)
			c <- NewResponseErr(req, err)
			return
		}

		resp := NewResponse(req)
		resp.Msg = in
		resp.Rtt = rtt

		c <- resp
	}()

	return c
}

func (r *Resolver) Exchange(ctx context.Context, reqs ...*Request) <-chan *Response {
	cs := []<-chan *Response{}

	for i := 0; i < len(reqs); i++ {
		c := r.Resolve(reqs[i])
		cs = append(cs, c)
	}

	return r.merge(ctx, cs...)
}

func (r *Resolver) merge(ctx context.Context, cs ...<-chan *Response) <-chan *Response {
	var wg sync.WaitGroup
	out := make(chan *Response)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c or done is closed, then calls
	// wg.Done.
	output := func(c <-chan *Response) {
		defer wg.Done()
		for resp := range c {
			select {
			case out <- resp:
			case <-ctx.Done():
				return
			}
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (r *Resolver) ResolveTypes(ctx context.Context, addr string, name string, types []uint16, options ...RequestOption) <-chan *Response {
	reqs := []*Request{}

	for i := 0; i < len(types); i++ {
		req := NewRequest(addr, name, types[i], options...)
		reqs = append(reqs, req)
	}

	return r.Exchange(ctx, reqs...)
}

func (r *Resolver) ResolveNames(ctx context.Context, addr string, type_ uint16, names []string, options ...RequestOption) <-chan *Response {
	reqs := []*Request{}

	for i := 0; i < len(names); i++ {
		req := NewRequest(addr, names[i], type_, options...)
		reqs = append(reqs, req)
	}

	return r.Exchange(ctx, reqs...)
}
