package resolv

import (
	"time"

	"github.com/miekg/dns"
)

type Response struct {
	Msg *dns.Msg
	Rtt time.Duration
	Err error
}
