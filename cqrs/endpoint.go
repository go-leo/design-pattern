package cqrs

import "context"

// Endpoint represents a single RPC method.
type Endpoint[Request any, Response any] interface {
	Invoke(ctx context.Context, request Request) (Response, error)
}

type EndpointFunc[Request any, Response any] func(ctx context.Context, request Request) (Response, error)

func (f EndpointFunc[Request, Response]) Invoke(ctx context.Context, request Request) (Response, error) {
	return f(ctx, request)
}

// QueryEndpoint convert QueryHandler to Endpoint
func QueryEndpoint[Q any, R any](h QueryHandler[Q, R]) Endpoint[Q, R] {
	return EndpointFunc[Q, R](func(ctx context.Context, request Q) (R, error) {
		return h.Handle(ctx, request)
	})
}

type emptyResponse struct{}

// CommandEndpoint convert CommandHandler to Endpoint
func CommandEndpoint[C any](h CommandHandler[C]) Endpoint[C, emptyResponse] {
	return EndpointFunc[C, emptyResponse](func(ctx context.Context, request C) (emptyResponse, error) {
		return emptyResponse{}, h.Handle(ctx, request)
	})
}
