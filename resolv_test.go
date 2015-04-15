package resolv_test

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/vitalie/resolv"
	"golang.org/x/net/context"
)

var nss = []string{
	"ns1.luadns.net",
	"ns2.luadns.net",
	"ns3.luadns.net",
	"ns4.luadns.net",
	"ns5.luadns.net",
}

// Simple resolve benchmark, run with:
//	go test -bench=.
func BenchmarkResolve(b *testing.B) {
	r := resolv.NewResolver()
	req := resolv.NewRequest("ns1.luadns.net", "cherpec.com", dns.TypeA)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp := <-r.Resolve(req)
		if resp.Err != nil {
			b.Fatal(resp.Err)
		}
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

func TestTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	r := resolv.NewResolver()
	req := resolv.NewRequest("ns1.luadns.net", "google.com", dns.TypeA)
	resp := <-r.Resolve(req)
	if resp.Err != nil {
		err, ok := resp.Err.(*resolv.DNSError)
		if !ok {
			t.Errorf("expected DNSError got: %v", resp.Err)
		}

		if !err.Timeout() {
			t.Errorf("expected timeout got: %v", resp.Err)
		}
	}
}

func TestFactoryFromNames(t *testing.T) {
	r := resolv.NewResolver()

	// Query multiple names
	fact := resolv.NewRequestFactory()
	reqs := fact.FromNames("ns1.luadns.net", dns.TypeA, "cherpec.com", "www.cherpec.com")

	n := 0
	c := r.FanIn(context.Background(), reqs...)
	for resp := range c {
		if resp.Err != nil {
			t.Error(resp.Err)
		}
		n++
	}

	if n != len(reqs) {
		t.Errorf("responses missmatch: %v != %v", n, len(reqs))
	}
}

func TestFactoryFromTypes(t *testing.T) {
	r := resolv.NewResolver()

	// Query multiple types
	fact := resolv.NewRequestFactory()
	reqs := fact.FromTypes("ns1.luadns.net", "cherpec.com", dns.TypeA, dns.TypeNS, dns.TypeMX)

	n := 0
	c := r.FanIn(context.Background(), reqs...)
	for resp := range c {
		if resp.Err != nil {
			t.Error(resp.Err)
		}
		n++
	}

	if n != len(reqs) {
		t.Errorf("responses missmatch: %v != %v", n, len(reqs))
	}
}

func TestDelegation(t *testing.T) {
	rs := resolv.NewResolver()
	it := resolv.NewDelegation(rs)

	r1 := <-it.Resolve(context.Background(), ".")
	if r1.Err != nil {
		t.Fatal(r1.Err)
	}

	r2 := <-it.Resolve(context.Background(), "com.")
	if r2.Err != nil {
		t.Fatal(r2.Err)
	}

	r3 := <-it.Resolve(context.Background(), "cherpec.com.")
	if r3.Err != nil {
		t.Fatal(r3.Err)
	}
}

func TestContext(t *testing.T) {
	rs := resolv.NewResolver()
	it := resolv.NewDelegation(rs)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	r := <-it.Resolve(ctx, "cherpec.com")
	if r.Err == nil {
		t.Fatal("expecting timeout got:", r)
	}
}
