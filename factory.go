package resolv

// RequestFactory creates multiple requests using custom options.
type RequestFactory struct {
	opts []RequestOption
}

// NewRequestFactory creates a factory using custom options.
func NewRequestFactory(opts ...RequestOption) *RequestFactory {
	return &RequestFactory{
		opts: opts,
	}
}

// FromNames creates multiple requests using different names.
func (f *RequestFactory) FromNames(addr string, type_ uint16, names ...string) []*Request {
	var reqs []*Request

	for _, name := range names {
		r := NewRequest(addr, name, type_, f.opts...)
		reqs = append(reqs, r)
	}

	return reqs
}

// FromTypes creates multiple requests using different types.
func (f *RequestFactory) FromTypes(addr, name string, types ...uint16) []*Request {
	var reqs []*Request

	for _, type_ := range types {
		r := NewRequest(addr, name, type_, f.opts...)
		reqs = append(reqs, r)
	}

	return reqs
}
