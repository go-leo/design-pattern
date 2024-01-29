package cqrs

import (
	"context"
	"github.com/go-leo/design-pattern/endpoint"
)

// CommandEndpoint convert CommandHandler to Endpoint
func CommandEndpoint[Command any](h CommandHandler[Command]) endpoint.Endpoint[Command, struct{}] {
	return endpoint.Func[Command, struct{}](func(ctx context.Context, request Command) (struct{}, error) {
		return struct{}{}, h.Handle(ctx, request)
	})
}

// QueryEndpoint convert QueryHandler to Endpoint
func QueryEndpoint[Query any, Result any](h QueryHandler[Query, Result]) endpoint.Endpoint[Query, Result] {
	return endpoint.Func[Query, Result](func(ctx context.Context, request Query) (Result, error) {
		return h.Handle(ctx, request)
	})
}
