package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"luadns.com/resolv"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: delegation <domain>\n")
	flag.PrintDefaults()
	os.Exit(2)
}

var verbose bool

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable debug messages")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Domain name is missing.")
		os.Exit(1)
	}

	r := resolv.NewResolver()
	d := resolv.NewDelegation(r)
	d.Verbose = verbose

	resp, err := d.Resolve(context.Background(), args[0])
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(resp)
}
