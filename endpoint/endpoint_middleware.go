package endpoint

// Middleware allows us to write something like decorators to Endpoint.
// It can execute something before invoke or after.
type Middleware[Request any, Response any] interface {
	// Decorate wraps the underlying Endpoint, adding some functionality.
	Decorate(handler Endpoint[Request, Response]) Endpoint[Request, Response]
}

// The MiddlewareFunc type is an adapter to allow the use of ordinary functions as Middleware.
// If f is a function with the appropriate signature, MiddlewareFunc(f) is a Middleware that calls f.
type MiddlewareFunc[Request any, Response any] func(handler Endpoint[Request, Response]) Endpoint[Request, Response]

// Decorate call f(endpoint).
func (f MiddlewareFunc[Request, Response]) Decorate(handler Endpoint[Request, Response]) Endpoint[Request, Response] {
	return f(handler)
}

// Chain decorates the given Endpoint with all middlewares.
func Chain[Q any, R any](handler Endpoint[Q, R], middlewares ...Middleware[Q, R]) Endpoint[Q, R] {
	var chain Endpoint[Q, R]
	chain = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i].Decorate(chain)
	}
	return chain
}
