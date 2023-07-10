package cqrs

// QueryMiddleware allows us to write something like decorators to QueryHandler.
// It can execute something before Handle or after.
type QueryMiddleware[Q any, R any] interface {
	// Decorate wraps the underlying Command, adding some functionality.
	Decorate(QueryHandler[Q, R]) QueryHandler[Q, R]
}

// The QueryMiddlewareFunc type is an adapter to allow the use of ordinary functions as QueryMiddleware.
// If f is a function with the appropriate signature, QueryMiddlewareFunc(f) is a QueryMiddleware that calls f.
type QueryMiddlewareFunc[Q any, R any] func(QueryHandler[Q, R]) QueryHandler[Q, R]

// Decorate call f(cmd).
func (f QueryMiddlewareFunc[Q, R]) Decorate(handler QueryHandler[Q, R]) QueryHandler[Q, R] {
	return f(handler)
}

// ChainQueryHandler decorates the given QueryHandler with all middlewares.
func ChainQueryHandler[Q any, R any](handler QueryHandler[Q, R], middlewares ...QueryMiddleware[Q, R]) QueryHandler[Q, R] {
	var chain QueryHandler[Q, R]
	chain = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i].Decorate(chain)
	}
	return chain
}
