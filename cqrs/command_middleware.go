package cqrs

// CommandMiddleware allows us to write something like decorators to CommandHandler.
// It can execute something before Handle or after.
type CommandMiddleware[C any] interface {
	// Decorate wraps the underlying Command, adding some functionality.
	Decorate(handler CommandHandler[C]) CommandHandler[C]
}

// The CommandMiddlewareFunc type is an adapter to allow the use of ordinary functions as CommandMiddleware.
// If f is a function with the appropriate signature, CommandMiddlewareFunc(f) is a CommandMiddleware that calls f.
type CommandMiddlewareFunc[C any] func(handler CommandHandler[C]) CommandHandler[C]

// Decorate call f(cmd).
func (f CommandMiddlewareFunc[C]) Decorate(handler CommandHandler[C]) CommandHandler[C] {
	return f(handler)
}

// ChainCommandHandler decorates the given CommandHandler with all middlewares.
func ChainCommandHandler[C any](handler CommandHandler[C], middlewares ...CommandMiddleware[C]) CommandHandler[C] {
	var chain CommandHandler[C]
	chain = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i].Decorate(chain)
	}
	return chain
}
