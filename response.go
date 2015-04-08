package resolv

import (
	"time"

	"github.com/miekg/dns"
)

type Response struct {
	Req *Request
	Msg *dns.Msg
	Rtt time.Duration
	Err error
}

func NewResponse(req *Request) *Response {
	r := &Response{
		Req: req,
	}

	return r
}

func NewResponseErr(req *Request, err error) *Response {
	r := NewResponse(req)
	r.Err = err
	return r
}

func (r *Response) Addr() string {
	if r.Req == nil {
		return "<nil>"
	}

	return r.Req.Addr
}
