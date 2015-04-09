package resolv

import (
	"fmt"
	"log"
	"strings"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

const (
	MaxIterations = 16
)

type Delegation struct {
	Path []*Response
	Err  error
}

type DelegIter struct {
	Verbose bool
	rs      *Resolver
}

func NewDelegIter(r *Resolver) *DelegIter {
	return &DelegIter{rs: r}
}

func (it *DelegIter) Resolve(ctx context.Context, domain string) <-chan *Delegation {
	out := make(chan *Delegation, 1)

	go func() {
		defer close(out)
		path, err := it.run(ctx, domain, RootServers...)
		out <- &Delegation{Path: path, Err: err}
	}()

	return out
}

func (it *DelegIter) run(ctx context.Context, domain string, nss ...string) ([]*Response, error) {
	var ns string

	skip := map[string]bool{}
	fqdn := dns.Fqdn(strings.ToLower(domain))
	path := []*Response{}

	for i := 0; len(nss) > 0; i++ {
		if i > MaxIterations {
			return nil, fmt.Errorf("iterator: max iterations reached")
		}

		ns, nss = peekRandom(nss)
		skip[ns] = true

		req := NewRequest(ns, fqdn, dns.TypeNS)
		c := it.rs.Resolve(req)
		select {
		case resp := <-c:
			if it.Verbose {
				log.Println("iterator: servers=", nss)
				log.Println("iterator: ===>", resp.Addr(), resp)
			}

			if resp.Err != nil {
				if err, ok := resp.Err.(*DNSError); ok {
					switch {
					case err.NameError():
						return nil, err
					default:
						continue
					}
				}
				return nil, resp.Err
			}

			path = append(path, resp)
			if _, ok := it.Search(resp.Msg.Answer, fqdn); ok {
				return path, nil
			}

			if referals, ok := it.Search(resp.Msg.Ns, fqdn); ok {
				return path, nil
			} else {
				if len(referals) > 0 {
					nss = []string{}
					for _, ns := range referals {
						if _, ok := skip[ns]; !ok {
							nss = append(nss, ns)
						}
					}
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("iterator: no more servers to try")
}

func (it *DelegIter) Search(section []dns.RR, domain string) ([]string, bool) {
	nss := []string{}

	for _, i := range section {
		switch i.(type) {
		case *dns.NS:
			rr := i.(*dns.NS)
			nm := strings.ToLower(rr.Header().Name)

			// DelegIter found.
			if nm == domain {
				return nil, true
			}

			// Collect referals.
			if strings.HasSuffix(domain, nm) {
				nss = append(nss, strings.ToLower(rr.Ns))
			}
		default:
			log.Println("iterator: bad RR type", i, "for", domain)
		}
	}

	return nss, false
}
