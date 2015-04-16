package resolv

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// Resolver represents an resolver, an application
// can have multiple resolvers.
type Resolver struct {
}

// NewResolver creates a new resolver with default options.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve issue the DNS request returning immediately,
// it returns the response through a channel which is
// closed automatically when the request is finished.
func (r *Resolver) Resolve(req *Request) <-chan *Response {
	// Buffered channel to avoid goroutine leaking.
	c := make(chan *Response, 1)

	// Launch a gorouting which will resolve the request
	// and will return the result when it's ready.
	go func() {
		defer func() {
			// Handle unexpected errors from the DNS librariy.
			if err := recover(); err != nil {
				log.Println("resolv: unexpected error:", err)
				c <- &Response{Err: fmt.Errorf("resolve: %v", err)}
			}
			close(c)
		}()

		// Prepare the message.
		m := new(dns.Msg)
		m.Id = dns.Id()
		m.RecursionDesired = req.Recurse
		m.Question = make([]dns.Question, 1)
		m.Question[0] = dns.Question{req.Name, req.Type, req.Class}

		cli := new(dns.Client)
		cli.Net = req.Mode

		// Issue synchronous request.
		in, rtt, err := cli.Exchange(m, req.Addr)
		if err != nil {
			// Timeout.
			if nerr, ok := err.(*net.OpError); ok && nerr.Timeout() {
				err := NewDNSError("timeout", req)
				err.IsTimeout = true
				c <- &Response{Err: err}
				return
			}

			// Other error.
			c <- &Response{Err: err}
			return
		}

		// Check the RCODE from the message.
		if in.Rcode != dns.RcodeSuccess {
			err := NewDNSError(
				dns.RcodeToString[in.Rcode],
				req,
			)

			if in.Rcode == dns.RcodeNameError {
				err.IsNameError = true
			}

			c <- &Response{Err: err}
			return
		}

		// Handle trruncated messages.
		if in.MsgHdr.Truncated {
			err := NewDNSError(
				"truncated",
				req,
			)
			c <- &Response{Err: err}
			return
		}

		resp := &Response{
			Req: req,
			Msg: in,
			Rtt: rtt,
		}

		c <- resp
	}()

	return c
}

// FanIn issues multiple requests and serializes the responses through the returned channel.
func (r *Resolver) FanIn(ctx context.Context, reqs ...*Request) <-chan *Response {
	cs := []<-chan *Response{}

	for i := 0; i < len(reqs); i++ {
		c := r.Resolve(reqs[i])
		cs = append(cs, c)
	}

	return r.merge(ctx, cs...)
}

// merge merges multiple channels into a single channel.
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
				out <- &Response{Err: ctx.Err()}
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
