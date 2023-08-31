package cqrs

import (
	"context"
	"github.com/go-leo/design-pattern/endpoint"
)

// CommandEndpoint convert CommandHandler to Endpoint
func CommandEndpoint[C any](h CommandHandler[C]) endpoint.Endpoint[C, struct{}] {
	return endpoint.EndpointFunc[C, struct{}](func(ctx context.Context, request C) (struct{}, error) {
		return struct{}{}, h.Handle(ctx, request)
	})
}

// QueryEndpoint convert QueryHandler to Endpoint
func QueryEndpoint[Q any, R any](h QueryHandler[Q, R]) endpoint.Endpoint[Q, R] {
	return endpoint.EndpointFunc[Q, R](func(ctx context.Context, request Q) (R, error) {
		return h.Handle(ctx, request)
	})
}
