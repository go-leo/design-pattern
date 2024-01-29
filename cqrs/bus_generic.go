package cqrs

import (
	"context"
	"fmt"
	"github.com/go-leo/gox/syncx/chanx"
)

type GenericBus[Result any] interface {
	// RegisterCommand register CommandHandler
	RegisterCommand(handler any) error
	// RegisterQuery register QueryHandler
	RegisterQuery(handler any) error

	// Exec sync execute command
	Exec(ctx context.Context, cmd any) error
	// Query sync query Query
	Query(ctx context.Context, q any) (Result, error)

	// AsyncExec async execute command
	AsyncExec(ctx context.Context, cmd any) <-chan error
	// AsyncQuery async query Query
	AsyncQuery(ctx context.Context, q any) (<-chan Result, <-chan error)

	// Close bus gracefully
	Close(ctx context.Context) error
}

type genericBus[Result any] struct {
	bus Bus
}

// RegisterCommand register CommandHandler
func (bus *genericBus[Result]) RegisterCommand(handler any) error {
	return bus.bus.RegisterCommand(handler)
}

// RegisterQuery register QueryHandler
func (bus *genericBus[Result]) RegisterQuery(handler any) error {
	return bus.RegisterQuery(handler)
}

// Exec sync execute command
func (bus *genericBus[Result]) Exec(ctx context.Context, cmd any) error {
	return bus.bus.Exec(ctx, cmd)
}

// Query sync query Query
func (bus *genericBus[Result]) Query(ctx context.Context, q any) (Result, error) {
	var r Result
	res, err := bus.bus.Query(ctx, q)
	if err != nil {
		return r, err
	}
	r, ok := res.(Result)
	if ok {
		return r, nil
	}
	return r, ConvertError{Res: res}
}

// AsyncExec async execute command
func (bus *genericBus[Result]) AsyncExec(ctx context.Context, cmd any) <-chan error {
	return bus.bus.AsyncExec(ctx, cmd)
}

// AsyncQuery async query Query
func (bus *genericBus[Result]) AsyncQuery(ctx context.Context, q any) (<-chan Result, <-chan error) {
	rC := make(chan Result, 1)
	resC, errC := bus.bus.AsyncQuery(ctx, q)
	go func() {
		defer close(rC)
		res, ok := <-resC
		if ok {
			if result, ok := res.(Result); ok {
				rC <- result
				return
			} else {
				convErrC := make(chan error, 1)
				convErrC <- ConvertError{Res: res}
				close(convErrC)
				errC = chanx.Combine(errC, convErrC)
				return
			}
		}
	}()
	return rC, errC
}

// Close bus gracefully
func (bus *genericBus[Result]) Close(ctx context.Context) error {
	return bus.bus.Close(ctx)
}

type ConvertError struct {
	Res any
}

func (c ConvertError) Error() string {
	return fmt.Errorf("failed to convert query result, %v", c.Res).Error()
}

func NewGenericBus[Result any](bus Bus) GenericBus[Result] {
	return &genericBus[Result]{bus: bus}
}
