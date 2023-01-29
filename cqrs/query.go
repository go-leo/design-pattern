package cqrs

import "context"

type QueryHandler[Q any, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

// The QueryHandlerFunc type is an adapter to allow the use of ordinary functions as QueryHandler.
// If f is a function with the appropriate signature, QueryHandlerFunc(f) is a QueryHandler that calls f.
type QueryHandlerFunc[Q any, R any] func(ctx context.Context, q Q) (R, error)

// Handle calls f(ctx).
func (f QueryHandlerFunc[Q, R]) Handle(ctx context.Context, q Q) (R, error) {
	return f(ctx, q)
}

// QueryHandlerMiddleware allows us to write something like decorators to QueryHandler.
// It can execute something before Handle or after.
type QueryHandlerMiddleware[Q any, R any] interface {
	// Decorate wraps the underlying Command, adding some functionality.
	Decorate(QueryHandler[Q, R]) QueryHandler[Q, R]
}

// The QueryHandlerMiddlewareFunc type is an adapter to allow the use of ordinary functions as QueryHandlerMiddleware.
// If f is a function with the appropriate signature, QueryHandlerMiddlewareFunc(f) is a QueryHandlerMiddleware that calls f.
type QueryHandlerMiddlewareFunc[Q any, R any] func(QueryHandler[Q, R]) QueryHandler[Q, R]

// Decorate call f(cmd).
func (f QueryHandlerMiddlewareFunc[Q, R]) Decorate(handler QueryHandler[Q, R]) QueryHandler[Q, R] {
	return f(handler)
}

// ChainQueryHandler decorates the given QueryHandler with all middlewares.
func ChainQueryHandler[Q any, R any](handler QueryHandler[Q, R], middlewares ...QueryHandlerMiddleware[Q, R]) QueryHandler[Q, R] {
	var chain QueryHandler[Q, R]
	chain = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i].Decorate(chain)
	}
	return chain
}
