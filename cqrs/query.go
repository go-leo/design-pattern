package cqrs

import (
	"context"
)

// QueryHandler is a query handler that to queries to read data.
// Queries never modify the database.
// A query returns a DTO that does not encapsulate any domain knowledge.
type QueryHandler[Query any, Result any] interface {
	Handle(ctx context.Context, q Query) (Result, error)
}

// The QueryHandlerFunc type is an adapter to allow the use of ordinary functions as QueryHandler.
type QueryHandlerFunc[Query any, Result any] func(ctx context.Context, q Query) (Result, error)

// Handle calls f(ctx).
func (f QueryHandlerFunc[Query, Result]) Handle(ctx context.Context, q Query) (Result, error) {
	return f(ctx, q)
}

// NoopQuery is an QueryHandler that does nothing and returns a nil error.
type NoopQuery[Query any, Result any] struct{}

func (NoopQuery[Query, Result]) Handle(context.Context, Query) (Result, error) {
	return *(new(Result)), nil
}
