package resolv_test

import (
	"testing"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"luadns.com/resolv"
)

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
	req = resolv.NewRequest("ns5.linode.com", "version.bind", dns.TypeTXT, resolv.SetCHAOSClass)
	resp = <-r.Resolve(req)
	if resp.Err != nil {
		t.Fatal(resp.Err)
	}
}

func TestFanIn(t *testing.T) {
	r := resolv.NewResolver()

	// Query multiple types
	types := []uint16{dns.TypeA, dns.TypeNS, dns.TypeMX}
	c1 := r.ResolveTypes(context.Background(), resolv.ProtoUDP, "8.8.8.8", "cherpec.com", types...)
	if len(c1) != len(types) {
		t.Errorf("responses missmatch: %v != %v", len(c1), len(types))
	}

	// Query multiple names
	names := []string{"cherpec.com", "www.cherpec.com"}
	c2 := r.ResolveNames(context.Background(), resolv.ProtoUDP, "8.8.8.8", dns.TypeA, names...)
	if len(c2) != len(names) {
		t.Errorf("responses missmatch: %v != %v", len(c2), len(names))
	}
}
