package resolv

import (
	"github.com/miekg/dns"
)

type DNSError struct {
	Err   string
	Name  string
	Type  uint16
	Class uint16
	Addr  string

	IsNameError bool
	IsTimeout   bool
}

func NewDNSError(err string, req *Request) *DNSError {
	e := &DNSError{
		Err:   err,
		Name:  req.Name,
		Type:  req.Type,
		Class: req.Class,
		Addr:  req.Addr,
	}

	return e
}

func (e *DNSError) NameError() bool { return e.IsNameError }
func (e *DNSError) Timeout() bool   { return e.IsTimeout }
func (e *DNSError) Temporary() bool { return e.IsTimeout }

func (e *DNSError) Error() string {
	if e == nil {
		return "<nil>"
	}

	s := "lookup " + e.Name
	s += " type " + dns.TypeToString[e.Type]
	s += " class " + dns.ClassToString[e.Class]

	if e.Addr != "" {
		s += " on " + e.Addr
	}

	s += ": " + e.Err
	return s
}
