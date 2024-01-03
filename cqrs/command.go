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

// The commandHandlerFunc type is an adapter to allow the use of ordinary functions as CommandHandler.
type commandHandlerFunc[C any] struct {
	f func(ctx context.Context, cmd C) error
}

// Handle calls f(ctx).
func (f commandHandlerFunc[C]) Handle(ctx context.Context, cmd C) error {
	return f.f(ctx, cmd)
}

func CommandHandlerFunc[C any](f func(ctx context.Context, cmd C) error) CommandHandler[C] {
	return commandHandlerFunc[C]{f: f}
}
