package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/vitalie/resolv"
	"golang.org/x/net/context"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: delegation <domain>\n")
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
		fmt.Println("Domain name is missing.")
		os.Exit(1)
	}

	rs := resolv.NewResolver()
	it := resolv.NewDelegation(rs)
	it.Debug = debug

	r := <-it.Resolve(context.Background(), args[0])
	if r.Err != nil {
		log.Fatalln(r.Err)
	}

	for _, resp := range r.Path {
		fmt.Println("=====>", resp.Addr())
		fmt.Println(resp.Msg)
	}
}
