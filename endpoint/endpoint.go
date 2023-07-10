package endpoint

import "context"

// Endpoint represents a single RPC method.
type Endpoint[Req any, Resp any] interface {
	Invoke(ctx context.Context, request Req) (Resp, error)
}

// The EndpointFunc type is an adapter to allow the use of ordinary functions as Endpoint.
// If f is a function with the appropriate signature, EndpointFunc(f) is a Endpoint that calls f.
type EndpointFunc[Req any, Resp any] func(ctx context.Context, request Req) (Resp, error)

// Invoke calls f(ctx).
func (f EndpointFunc[Req, Resp]) Invoke(ctx context.Context, request Req) (Resp, error) {
	return f(ctx, request)
}

// Noop is an endpoint that does nothing and returns a nil error.
type Noop[Req any, Resp any] struct{}

func (Noop[Req, Resp]) Invoke(context.Context, Req) (Resp, error) {
	var resp Resp
	return resp, nil
}
