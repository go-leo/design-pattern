package cqrs

import (
	"context"
	"github.com/go-leo/design-pattern/endpoint"
)

type emptyResponse struct{}

// CommandEndpoint convert CommandHandler to Endpoint
func CommandEndpoint[C any](h CommandHandler[C]) endpoint.Endpoint[C, emptyResponse] {
	return endpoint.EndpointFunc[C, emptyResponse](func(ctx context.Context, request C) (emptyResponse, error) {
		return emptyResponse{}, h.Handle(ctx, request)
	})
}

// QueryEndpoint convert QueryHandler to Endpoint
func QueryEndpoint[Q any, R any](h QueryHandler[Q, R]) endpoint.Endpoint[Q, R] {
	return endpoint.EndpointFunc[Q, R](func(ctx context.Context, request Q) (R, error) {
		return h.Handle(ctx, request)
	})
}
