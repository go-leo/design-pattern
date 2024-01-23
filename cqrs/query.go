package cqrs

import (
	"context"
)

// QueryHandler is a query handler that to queries to read data.
// Queries never modify the database.
// A query returns a DTO that does not encapsulate any domain knowledge.
type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

// The QueryHandlerFunc type is an adapter to allow the use of ordinary functions as QueryHandler.
type QueryHandlerFunc[Q any, R any] func(ctx context.Context, q Q) (R, error)

// Handle calls f(ctx).
func (f QueryHandlerFunc[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	return f(ctx, q)
}

// NoopQuery is an QueryHandler that does nothing and returns a nil error.
type NoopQuery[Q any, R any] struct{}

func (NoopQuery[Q, R]) Handle(context.Context, Q) (R, error) {
	var r R
	return r, nil
}
