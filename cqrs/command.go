package cqrs

import "context"

// CommandHandler is a command handler that to update data.
// Commands should be task-based, rather than data centric.
// Commands may be placed on a queue for asynchronous processing, rather than being processed synchronously.
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

// The CommandHandlerFunc type is an adapter to allow the use of ordinary functions as CommandHandler.
// If f is a function with the appropriate signature, CommandHandlerFunc(f) is a CommandHandler that calls f.
type CommandHandlerFunc[C any] func(ctx context.Context, cmd C) error

// Handle calls f(ctx).
func (f CommandHandlerFunc[C]) Handle(ctx context.Context, cmd C) error {
	return f(ctx, cmd)
}

// CommandHandlerMiddleware allows us to write something like decorators to CommandHandler.
// It can execute something before Handle or after.
type CommandHandlerMiddleware[C any] interface {
	// Decorate wraps the underlying Command, adding some functionality.
	Decorate(handler CommandHandler[C]) CommandHandler[C]
}

// The CommandHandlerMiddlewareFunc type is an adapter to allow the use of ordinary functions as CommandHandlerMiddleware.
// If f is a function with the appropriate signature, CommandHandlerMiddlewareFunc(f) is a CommandHandlerMiddleware that calls f.
type CommandHandlerMiddlewareFunc[C any] func(handler CommandHandler[C]) CommandHandler[C]

// Decorate call f(cmd).
func (f CommandHandlerMiddlewareFunc[C]) Decorate(handler CommandHandler[C]) CommandHandler[C] {
	return f(handler)
}

// ChainCommandHandler decorates the given CommandHandler with all middlewares.
func ChainCommandHandler[C any](handler CommandHandler[C], middlewares ...CommandHandlerMiddleware[C]) CommandHandler[C] {
	var chain CommandHandler[C]
	chain = handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i].Decorate(chain)
	}
	return chain
}
