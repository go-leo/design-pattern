package cqrs

import (
	"context"
	"github.com/go-leo/gox/syncx/chanx"
	"sync"
)

var globalBus Bus
var globalBusMutex sync.RWMutex

func SetBus(new Bus) Bus {
	globalBusMutex.Lock()
	defer globalBusMutex.Unlock()
	old := globalBus
	globalBus = new
	return old
}

func GetBus() Bus {
	globalBusMutex.RLock()
	defer globalBusMutex.RUnlock()
	return globalBus
}

func init() {
	globalBus = NewBus()
}

// RegisterCommand register CommandHandler
func RegisterCommand(handler any) error {
	return GetBus().RegisterCommand(handler)
}

// RegisterQuery register QueryHandler
func RegisterQuery(handler any) error {
	return GetBus().RegisterQuery(handler)
}

// Exec sync execute command
func Exec(ctx context.Context, cmd any) error {
	return GetBus().Exec(ctx, cmd)
}

// Query sync query Query
func Query[R any](ctx context.Context, q any) (R, error) {
	var r R
	res, err := GetBus().Query(ctx, q)
	if err != nil {
		return r, err
	}
	r, ok := res.(R)
	if ok {
		return r, nil
	}
	return r, ConvertError{Res: res}
}

// AsyncExec async execute command
func AsyncExec(ctx context.Context, cmd any) <-chan error {
	return GetBus().AsyncExec(ctx, cmd)
}

// AsyncQuery async query Query
func AsyncQuery[R any](ctx context.Context, q any) (<-chan R, <-chan error) {
	rC := make(chan R, 1)
	resC, errC := GetBus().AsyncQuery(ctx, q)
	go func() {
		defer close(rC)
		res, ok := <-resC
		if ok {
			if result, ok := res.(R); ok {
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
func Close(ctx context.Context) error {
	return GetBus().Close(ctx)
}

type ConvertError struct {
	Res any
}

func (c ConvertError) Error() string {
	return "failed to convert result"
}
