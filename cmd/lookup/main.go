package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/vitalie/resolv"
	"golang.org/x/net/context"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: lookup <domain> <type>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var debug bool

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.BoolVar(&debug, "debug", false, "Enable debug messages")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Name is missing.")
		os.Exit(1)
	}

	qname := args[0]
	qtype := dns.TypeA

	if len(args) == 2 {
		s := strings.ToUpper(args[1])
		if v, ok := dns.StringToType[s]; ok {
			qtype = v
		} else {
			fmt.Println("Invalid type:", s, ".")
			os.Exit(1)
		}
	}

	rs := resolv.NewResolver()
	it := resolv.NewIterator(rs)
	it.Debug = debug

	ctx := context.Background()
	for r := range it.Resolve(ctx, qname, qtype) {
		if r.Err != nil {
			log.Fatalln(r.Err)
		}

		fmt.Println("=====>", r.Addr)
		fmt.Println(r.Msg)
	}
}
