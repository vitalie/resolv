package resolv_test

import (
	"testing"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"luadns.com/resolv"
)

var nss = []string{
	"ns1.luadns.net",
	"ns2.luadns.net",
	"ns3.luadns.net",
	"ns4.luadns.net",
	"ns5.luadns.net",
}

func TestShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
}

func TestResolve(t *testing.T) {
	r := resolv.NewResolver()

	// UDP Mode
	req := resolv.NewRequest("ns1.luadns.net", "cherpec.com", dns.TypeA)
	resp := <-r.Resolve(req)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}

	// TCP Mode
	req = resolv.NewRequest("ns1.luadns.net", "cherpec.com", dns.TypeA, resolv.SetTCPMode)
	resp = <-r.Resolve(req)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}

	// CHAOS class
	req = resolv.NewRequest("ns1.linode.com", "version.bind", dns.TypeTXT, resolv.SetCHAOSClass)
	resp = <-r.Resolve(req)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
}

func TestResolveTypes(t *testing.T) {
	r := resolv.NewResolver()

	// Query multiple types
	types := []uint16{dns.TypeA, dns.TypeNS, dns.TypeMX}
	c := r.ResolveTypes(context.Background(), "8.8.8.8", "cherpec.com", types)

	n := 0
	for resp := range c {
		if resp.Err != nil {
			t.Fatal(resp.Err)
		}
		n++
	}

	if n != len(types) {
		t.Errorf("responses missmatch: %v != %v", n, len(types))
	}
}

func TestResolveNames(t *testing.T) {
	r := resolv.NewResolver()

	// Query multiple names
	names := []string{"cherpec.com", "www.cherpec.com"}
	c := r.ResolveNames(context.Background(), "8.8.8.8", dns.TypeA, names)

	n := 0
	for resp := range c {
		if resp.Err != nil {
			t.Fatal(resp.Err)
		}
		n++
	}

	if n != len(names) {
		t.Errorf("responses missmatch: %v != %v", n, len(names))
	}
}
