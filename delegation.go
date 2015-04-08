package resolv

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const (
	MaxIterations = 16
)

type Iterator func(ctx context.Context, domain string, nss []string, n int, options ...RequestOption) (*Response, error)

type Delegation struct {
	resolver *Resolver
}

func NewDelegation(r *Resolver) *Delegation {
	return &Delegation{resolver: r}
}

func (r *Delegation) Resolve(ctx context.Context, domain string, options ...RequestOption) <-chan *Response {
	out := make(chan *Response, 1)
	used := map[string]bool{}

	var iterator Iterator
	iterator = func(ctx context.Context, domain string, nss []string, n int, options ...RequestOption) (*Response, error) {
		if n == 0 {
			return nil, fmt.Errorf("iterator: max iterations reached")
		}

		if len(nss) == 0 {
			return nil, fmt.Errorf("iterator: no more servers to try")
		}

		ns, nss := r.PeekRandom(nss)
		if _, ok := used[ns]; ok {
			return iterator(ctx, domain, nss, n-1, options...)
		}
		used[ns] = true

		req := NewRequest(ns, domain, dns.TypeNS)
		c := r.resolver.Resolve(req)
		select {
		case resp := <-c:
			log.Println("iterator:", resp)

			if resp.Err != nil {
				return iterator(ctx, domain, nss, n-1, options...)
			}

			if r.Found(resp.Msg, domain) {
				return resp, nil
			}

			for _, ns := range r.Referals(resp.Msg, domain) {
				if _, ok := used[ns]; !ok {
					nss = append(nss, ns)
				}
			}
			return iterator(ctx, domain, nss, n-1, options...)
		case <-ctx.Done():
			// FIXME: ...
			return nil, nil
		}
	}

	go func() {
		defer close(out)
		// FIXME: ...
		resp, _ := iterator(ctx, domain, RootServers, MaxIterations, options...)
		out <- resp
	}()

	return out
}

func (r *Delegation) Found(msg *dns.Msg, domain string) bool {
	fqdn := dns.Fqdn(domain)

	// Anser section
	if len(msg.Answer) > 0 {
		for _, i := range msg.Answer {
			switch i.(type) {
			case *dns.NS:
				rr := i.(*dns.NS)
				if rr.Header().Name == fqdn {
					return true
				}
			default:
				log.Println("iterator: bad RR type", i, "for", domain)
			}
		}
	}

	return false
}

func (r *Delegation) Referals(msg *dns.Msg, domain string) []string {
	fqdn := dns.Fqdn(domain)

	// Authority section
	nss := []string{}
	for _, i := range msg.Ns {
		switch i.(type) {
		case *dns.NS:
			rr := i.(*dns.NS)
			if strings.HasSuffix(fqdn, rr.Header().Name) {
				nss = append(nss, rr.Ns)
			}
		default:
			log.Println("iterator: bad RR type", i, "for", domain)
		}
	}

	return nss
}

func (r *Delegation) PeekRandom(nss []string) (string, []string) {
	i := rand.Intn(len(nss) - 1)
	return nss[i], append(nss[:i], nss[i+1:]...)
}
