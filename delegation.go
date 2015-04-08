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
	Verbose  bool
	resolver *Resolver
}

func NewDelegation(r *Resolver) *Delegation {
	return &Delegation{resolver: r}
}

func (r *Delegation) Resolve(ctx context.Context, domain string, options ...RequestOption) (*Response, error) {
	used := map[string]bool{}
	fqdn := dns.Fqdn(strings.ToLower(domain))

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
			if r.Verbose {
				log.Println("iterator:", resp)
			}

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

	return iterator(ctx, fqdn, RootServers, MaxIterations, options...)
}

func (r *Delegation) Found(msg *dns.Msg, domain string) bool {
	// Anser section
	if len(msg.Answer) > 0 {
		for _, i := range msg.Answer {
			switch i.(type) {
			case *dns.NS:
				rr := i.(*dns.NS)
				nm := strings.ToLower(rr.Header().Name)
				if nm == domain {
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
	// Authority section
	nss := []string{}
	for _, i := range msg.Ns {
		switch i.(type) {
		case *dns.NS:
			rr := i.(*dns.NS)
			nm := strings.ToLower(rr.Header().Name)
			if strings.HasSuffix(domain, nm) {
				nss = append(nss, strings.ToLower(rr.Ns))
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
