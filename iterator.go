package resolv

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// Iterator represents an iterative DNS query.
type Iterator struct {
	Debug bool
	rs    *Resolver
}

// NewIterator initializes an Iterator structure.
func NewIterator(r *Resolver) *Iterator {
	return &Iterator{rs: r}
}

// LookupIPv4 looks up the IPv4 addresses for the host.
func (it *Iterator) LookupIPv4(ctx context.Context, host string) ([]net.IP, error) {
	fqdn := dns.Fqdn(strings.ToLower(host))

	var last *Response
	for r := range it.Resolve(ctx, fqdn, dns.TypeA) {
		if r.Err != nil {
			return nil, r.Err
		}
		last = r
	}

	var ips []net.IP
	for _, i := range last.Msg.Answer {
		switch i.(type) {
		case (*dns.A):
			rr := i.(*dns.A)
			ips = append(ips, rr.A)
		default:
			return nil, fmt.Errorf("iterator: unexpected record: %v", i)
		}
	}

	return ips, nil
}

// Resolve looks up the name starting from root servers following referals.
func (it *Iterator) Resolve(ctx context.Context, name string, type_ uint16) <-chan *Response {
	return it.run(ctx, name, type_, 0, 0, map[string]bool{}, RootServers...)
}

func (it *Iterator) run(ctx context.Context, name string, type_ uint16, depth, i int, skip map[string]bool, nss ...string) <-chan *Response {
	out := make(chan *Response, MaxDepth)
	defer close(out)

	if depth > MaxDepth {
		out <- &Response{Err: fmt.Errorf("iterator: max depth reached")}
		return out
	}

	for len(nss) > 0 {
		var ns string

		i++
		if i > MaxIterations {
			out <- &Response{Err: fmt.Errorf("iterator: max iterations reached")}
			return out
		}

		// Peek random name server and mark it as used.
		ns, nss = peekRandom(nss)
		if _, ok := skip[ns]; ok {
			continue
		}
		skip[ns] = true

		// Issue DNS query.
		fqdn := dns.Fqdn(strings.ToLower(name))
		c := it.rs.Resolve(
			NewRequest(ns, fqdn, type_),
		)

		select {
		case resp := <-c:
			if it.Debug {
				log.Println("iterator: servers=", nss)
				log.Println("iterator: ===>", resp.Addr(), resp)
			}

			if resp.Err != nil {
				// Ignore DNS errors, stop if error is NXDOMAIN.
				if err, ok := resp.Err.(*DNSError); ok {
					switch {
					case err.NameError():
						out <- &Response{Err: err}
						return out
					default:
						continue
					}
				}
				out <- resp
				return out
			}

			if referals, cname, ok := it.lookup(resp.Msg, fqdn); ok {
				if cname != "" {
					return it.run(ctx, cname, type_, depth+1, i, map[string]bool{}, RootServers...)
				} else {
					out <- resp
					return out
				}
			} else {
				if len(referals) > 0 {
					return it.run(ctx, name, type_, depth+1, i, skip, referals...)
				}
			}
		case <-ctx.Done():
			out <- &Response{Err: ctx.Err()}
			return out
		}
	}

	out <- &Response{Err: fmt.Errorf("iterator: no more servers to try")}
	return out
}

func (it *Iterator) lookup(msg *dns.Msg, fqdn string) ([]string, string, bool) {
	cname := ""
	found := false

	// Look for fqdn in the Answer section.
	for _, i := range msg.Answer {
		nm := strings.ToLower(i.Header().Name)
		if nm == fqdn {
			found = true
		}

		switch i.(type) {
		case (*dns.CNAME):
			rr := i.(*dns.CNAME)
			cname = dns.Fqdn(strings.ToLower(rr.Target))
		}
	}

	if found {
		return nil, cname, found
	}

	// If fqdn is not found then collect and return referals.
	var nss []string
	for _, i := range msg.Ns {
		switch i.(type) {
		case *dns.NS:
			rr := i.(*dns.NS)
			nm := strings.ToLower(rr.Header().Name)

			if strings.HasSuffix(fqdn, nm) {
				nss = append(nss, strings.ToLower(rr.Ns))
			}
		default:
			log.Println("iterator: bad RR type", i, "for", fqdn)
		}
	}

	return nss, cname, found
}
