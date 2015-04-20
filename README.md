# Resolv [![GoDoc](https://godoc.org/github.com/vitalie/resolv?status.svg)](http://godoc.org/github.com/vitalie/resolv) [![Build Status](https://travis-ci.org/vitalie/resolv.svg?branch=master)](https://travis-ci.org/vitalie/resolv)

A [context](https://godoc.org/golang.org/x/net/context) aware DNS library
based on [Miek](https://github.com/miekg/dns)'s library. It's goal to
offer a simple API to most common DNS operations while preserving access
to low level data.

## Installation

``` bash
$ go get -u github.com/vitalie/resolv
```

## Usage

``` go
package main

import (
    "fmt"
    "log"

    "github.com/miekg/dns"
    "golang.org/x/net/context"
    "github.com/vitalie/resolv"
)

func main() {
    r := resolv.NewResolver()

    // Issue an UDP request using default options.
    req := resolv.NewRequest("ns1.luadns.net", "cherpec.com", dns.TypeA)
    resp := <-r.Resolve(req)
    if resp.Err != nil {
        log.Fatalln("error:", resp.Err)
    }
    fmt.Println(resp)

    // Create a request factory.
    fact := resolv.NewRequestFactory()

    // Query multiple types.
    reqs := fact.FromTypes("ns1.luadns.net", "cherpec.com", dns.TypeA, dns.TypeNS, dns.TypeMX)
    for resp := range r.FanIn(context.Background(), reqs...) {
      fmt.Println(resp)
    }

    // Query multiple names.
    reqs = fact.FromNames("ns1.luadns.net", dns.TypeA, "cherpec.com", "www.cherpec.com")
    for resp := range r.FanIn(context.Background(), reqs...) {
      fmt.Println(resp)
    }
}
```

## TODO

- Cache responses
- Retry queries on timeout
- Don't query local resolver
- Switch to TCP on truncated responses

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
