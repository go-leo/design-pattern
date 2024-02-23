package cqrs

import (
	"context"
	"github.com/go-leo/gox/contextx"
	"github.com/go-leo/gox/errorx"
	"github.com/go-leo/gox/syncx"
	"reflect"
	"sync"
	"sync/atomic"
)

// Bus is a bus, register CommandHandler and QueryHandler, execute Command and query Query
type Bus interface {

	// RegisterCommand register CommandHandler
	RegisterCommand(handler any) error

	// RegisterQuery register QueryHandler
	RegisterQuery(handler any) error

	// Exec synchronously execute command
	Exec(ctx context.Context, cmd any) error

	// Query synchronously query Query
	Query(ctx context.Context, q any) (any, error)

	// AsyncExec asynchronously execute command
	AsyncExec(ctx context.Context, cmd any) <-chan error

	// AsyncQuery asynchronously query Query
	AsyncQuery(ctx context.Context, q any) (<-chan any, <-chan error)

	// Close bus gracefully
	Close(ctx context.Context) error
}

var _ Bus = (*bus)(nil)

type bus struct {
	commands   sync.Map
	queries    sync.Map
	wg         sync.WaitGroup
	inShutdown atomic.Bool // true when bus is in shutdown
	options    *option
}

func (b *bus) RegisterCommand(handler any) error {
	if handler == nil {
		return ErrHandlerNil
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()
	method, ok := handlerType.MethodByName("Handle")
	if !ok {
		return ErrUnimplemented
	}
	if method.Type.NumIn() != 3 {
		return ErrUnimplemented
	}
	if !method.Type.In(1).Implements(contextx.ContextType) {
		return ErrUnimplemented
	}
	if method.Type.NumOut() != 1 {
		return ErrUnimplemented
	}
	if !method.Type.Out(0).Implements(errorx.ErrorType) {
		return ErrUnimplemented
	}
	inType := method.Type.In(2)
	info := &handlerInfo{
		handlerVal:    handlerVal,
		handlerMethod: method,
		inType:        inType,
	}
	if _, loaded := b.commands.LoadOrStore(inType, info); loaded {
		return ErrRegistered
	}
	return nil
}

func (b *bus) RegisterQuery(handler any) error {
	if handler == nil {
		return ErrHandlerNil
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	handlerVal := reflect.ValueOf(handler)
	handlerType := handlerVal.Type()
	method, ok := handlerType.MethodByName("Handle")
	if !ok {
		return ErrUnimplemented
	}
	if method.Type.NumIn() != 3 {
		return ErrUnimplemented
	}
	if !method.Type.In(1).Implements(contextx.ContextType) {
		return ErrUnimplemented
	}
	if method.Type.NumOut() != 2 {
		return ErrUnimplemented
	}
	if !method.Type.Out(1).Implements(errorx.ErrorType) {
		return ErrUnimplemented
	}
	inType := method.Type.In(2)
	info := &handlerInfo{
		handlerVal:    handlerVal,
		handlerMethod: method,
		inType:        inType,
	}
	if _, loaded := b.queries.LoadOrStore(inType, info); loaded {
		return ErrRegistered
	}
	return nil
}

func (b *bus) Exec(ctx context.Context, cmd any) error {
	if cmd == nil {
		return ErrCommandNil
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	value, ok := b.commands.Load(reflect.TypeOf(cmd))
	if !ok {
		return ErrUnregistered
	}
	info := value.(*handlerInfo)
	resultValues := info.handlerMethod.Func.Call(
		[]reflect.Value{
			info.handlerVal,
			reflect.ValueOf(ctx),
			reflect.ValueOf(cmd),
		})
	err := resultValues[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

func (b *bus) Query(ctx context.Context, q any) (any, error) {
	if q == nil {
		return nil, ErrQueryNil
	}
	if b.shuttingDown() {
		return nil, ErrBusClosed
	}
	value, ok := b.queries.Load(reflect.TypeOf(q))
	if !ok {
		return nil, ErrUnregistered
	}
	info := value.(*handlerInfo)
	resultValues := info.handlerMethod.Func.Call(
		[]reflect.Value{
			info.handlerVal,
			reflect.ValueOf(ctx),
			reflect.ValueOf(q),
		})
	err := resultValues[1].Interface()
	if err != nil {
		return nil, err.(error)
	}
	return resultValues[0].Interface(), nil
}

func (b *bus) AsyncExec(ctx context.Context, cmd any) <-chan error {
	errC := make(chan error, 2)
	b.wg.Add(1)
	err := b.options.Pool.Go(func() {
		defer b.wg.Done()
		err := b.Exec(ctx, cmd)
		if err != nil {
			errC <- err
			return
		}
	})
	if err != nil {
		errC <- err
	}
	return errC
}

func (b *bus) AsyncQuery(ctx context.Context, q any) (<-chan any, <-chan error) {
	errC := make(chan error, 2)
	resC := make(chan any, 1)
	b.wg.Add(1)
	err := b.options.Pool.Go(func() {
		defer b.wg.Done()
		res, err := b.Query(ctx, q)
		if err != nil {
			errC <- err
			return
		}
		resC <- res
	})
	if err == nil {
		errC <- err
	}
	return resC, errC
}

func (b *bus) Close(ctx context.Context) error {
	if b.inShutdown.CompareAndSwap(false, true) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-syncx.WaitNotify(&b.wg):
			return nil
		}
	}
	return ErrBusClosed
}

func (b *bus) shuttingDown() bool {
	return b.inShutdown.Load()
}

type handlerInfo struct {
	handlerVal    reflect.Value
	handlerMethod reflect.Method
	inType        reflect.Type
}
