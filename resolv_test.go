package resolv_test

import (
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/vitalie/resolv"
	"golang.org/x/net/context"
)

var nssMap = map[string]bool{
	"ns1.luadns.net.": true,
	"ns2.luadns.net.": true,
	"ns3.luadns.net.": true,
	"ns4.luadns.net.": true,
	"ns5.luadns.net.": true,
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
	req = resolv.NewRequest("ns1.softlayer.com", "version.bind", dns.TypeTXT, resolv.SetCHAOSClass)
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

func TestContext(t *testing.T) {
	rs := resolv.NewResolver()
	it := resolv.NewIterator(rs)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	_, err := it.LookupIPv4(ctx, "cherpec.com.")
	if err == nil {
		t.Fatal("expecting timeout got:", err)
	}
}

func TestIterator(t *testing.T) {
	rs := resolv.NewResolver()
	it := resolv.NewIterator(rs)

	a4, err := it.LookupIPv4(context.Background(), "www.cherpec.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(a4) == 0 {
		t.Fatal("expecting IPv4 addresses, got", a4)
	}

	a6, err := it.LookupIPv6(context.Background(), "ns1.softlayer.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(a6) == 0 {
		t.Fatal("expecting IPv6 addresses, got", a6)
	}

	as, err := it.LookupIP(context.Background(), "ns1.softlayer.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(as) == 0 {
		t.Fatal("expecting IP4, IPv6 addresses, got", as)
	}

	nss, err := it.LookupNS(context.Background(), "cherpec.com")
	if err != nil {
		t.Fatal(err)
	}

	if len(nss) == 0 {
		t.Fatal("expecting NS records, got", nss)
	}

	for _, ns := range nss {
		if _, ok := nssMap[ns.Host]; !ok {
			t.Fatal("unexpected NS record", ns)
		}
	}

	srv, nss, err := it.Delegation(context.Background(), "cherpec.com")
	if err != nil {
		t.Fatal(err)
	}

	if srv == "" {
		t.Fatal("expecting server name, got", srv)
	}

	if len(nss) == 0 {
		t.Fatal("expecting NS records, got", nss)
	}

	for _, ns := range nss {
		if _, ok := nssMap[ns.Host]; !ok {
			t.Fatal("unexpected NS record", ns)
		}
	}
}
