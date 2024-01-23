package cqrs

import (
	"context"
)

// CommandHandler is a command handler that to update data.
// Commands should be task-based, rather than data centric.
// Commands may be placed on a queue for asynchronous processing, rather than being processed synchronously.
type CommandHandler[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

// The CommandHandlerFunc type is an adapter to allow the use of ordinary functions as CommandHandler.
type CommandHandlerFunc[C any] func(ctx context.Context, cmd C) error

// Handle calls f(ctx).
func (f CommandHandlerFunc[C]) Handle(ctx context.Context, cmd C) error {
	return f(ctx, cmd)
}

// NoopCommand is an CommandHandler that does nothing and returns a nil error.
type NoopCommand[Q any, R any] struct{}

func (NoopCommand[Q, R]) Handle(context.Context, Q) error {
	return nil
}
