package resolv

// RequestFactory creates multiple requests using opts.
type RequestFactory struct {
	opts []RequestOption
}

func NewRequestFactory(opts ...RequestOption) *RequestFactory {
	return &RequestFactory{
		opts: opts,
	}
}

func (f *RequestFactory) FromNames(addr string, type_ uint16, names ...string) []*Request {
	var reqs []*Request

	for _, name := range names {
		r := NewRequest(addr, name, type_, f.opts...)
		reqs = append(reqs, r)
	}

	return reqs
}

func (f *RequestFactory) FromTypes(addr, name string, types ...uint16) []*Request {
	var reqs []*Request

	for _, type_ := range types {
		r := NewRequest(addr, name, type_, f.opts...)
		reqs = append(reqs, r)
	}

	return reqs
}
