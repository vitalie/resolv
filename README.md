# Resolv

A [context](https://godoc.org/golang.org/x/net/context) aware DNS library
based on [Miek](https://github.com/miekg/dns)'s library. It's goal to
offer a simple API to most common DNS operations while preserving access
to low level data.

  - [![GoDoc](https://godoc.org/github.com/vitalie/resolv?status.svg)](http://godoc.org/github.com/vitalie/resolv)
  - [![Build Status](https://travis-ci.org/vitalie/resolv.svg?branch=master)](https://travis-ci.org/vitalie/resolv)

## Installation

``` bash
$ go get -u github.com/vitalie/resolv
```

## Usage

``` go
package main

import (
    "log"

    "github.com/miekg/dns"
    "github.com/vitalie/resolv"
)

func main() {
    r := resolv.NewResolver()

    // UDP Mode
    req := resolv.NewRequest("ns1.luadns.net", "cherpec.com", dns.TypeA)
    resp := <-r.Resolve(req)
    if resp.Err != nil {
        log.Fatalln("error:", resp.Err)
    }

    log.Println(resp)
}
```

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
