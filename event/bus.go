package event

import "context"

type Bus interface {
	// On adds a listener to the event bus.
	On(event any, listener any) error

	// Prepend adds the listener to the beginning of the listeners array.
	Prepend(event any, listener any) error

	// Once adds a one-time listener to the event bus.
	Once(event any, listener any) error

	// PrependOnce adds a one-time listener to the beginning of the listeners array.
	PrependOnce(event any, listener any) error

	// Off Removes the specified listener from the listeners array.
	Off(event any, listener any) error

	// Emit synchronously calls each of the listeners registered
	Emit(ctx context.Context, event any) error

	// AsyncEmit asynchronously calls each of the listeners registered
	AsyncEmit(ctx context.Context, event any) error

	// Close bus gracefully
	Close(ctx context.Context) error
}
