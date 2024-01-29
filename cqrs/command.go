package cqrs

import (
	"context"
)

// CommandHandler is a command handler that to update data.
// Commands should be task-based, rather than data centric.
// Commands may be placed on a queue for asynchronous processing, rather than being processed synchronously.
type CommandHandler[Command any] interface {
	Handle(ctx context.Context, cmd Command) error
}

// The CommandHandlerFunc type is an adapter to allow the use of ordinary functions as CommandHandler.
type CommandHandlerFunc[Command any] func(ctx context.Context, cmd Command) error

// Handle calls f(ctx).
func (f CommandHandlerFunc[Command]) Handle(ctx context.Context, cmd Command) error {
	return f(ctx, cmd)
}

// NoopCommand is an CommandHandler that does nothing and returns a nil error.
type NoopCommand[Command any] struct{}

func (NoopCommand[Command]) Handle(context.Context, Command) error { return nil }
