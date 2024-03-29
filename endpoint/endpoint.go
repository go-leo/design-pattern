package endpoint

import "context"

// Endpoint represents a single RPC method.
type Endpoint[Req any, Resp any] interface {
	Invoke(ctx context.Context, req Req) (Resp, error)
}

// The Func type is an adapter to allow the use of ordinary functions as Endpoint.
type Func[Req any, Resp any] func(ctx context.Context, req Req) (Resp, error)

// Invoke calls f(ctx).
func (f Func[Req, Resp]) Invoke(ctx context.Context, req Req) (Resp, error) {
	return f(ctx, req)
}

// Noop is an endpoint that does nothing and returns a nil error.
type Noop[Req any, Resp any] struct{}

func (Noop[Req, Resp]) Invoke(context.Context, Req) (Resp, error) {
	var resp Resp
	return resp, nil
}

type AnyEndpoint Endpoint[any, any]

type AnyEndpointFunc Func[any, any]
