package cqrs

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-leo/gox/contextx"
	"github.com/go-leo/gox/errorx"
	"github.com/go-leo/gox/syncx"
	"github.com/go-leo/gox/syncx/chanx"
	"github.com/go-leo/gox/syncx/gopher"
	"github.com/go-leo/gox/syncx/gopher/sample"
	"reflect"
	"sync"
	"sync/atomic"
)

var (
	// ErrHandlerNil CommandHandler or QueryHandler is nil
	ErrHandlerNil = errors.New("handler is nil")
	// ErrRegistered not register CommandHandler or QueryHandler
	ErrRegistered = errors.New("handler registered")
	// ErrCommandNil Command arg is nil
	ErrCommandNil = errors.New("command is nil")
	// ErrQueryNil Query arg is nil
	ErrQueryNil = errors.New("query is nil")
	// ErrUnimplemented handler is not implement CommandHandler or QueryHandler
	ErrUnimplemented = errors.New("handler is not implement CommandHandler or QueryHandler")
	// ErrBusClosed bus is closed
	ErrBusClosed = errors.New("bus is closed")
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
		return errors.New("handler unregistered")
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
		return nil, errors.New("handler unregistered")
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
	errC := make(chan error, 1)
	b.wg.Add(1)
	err := b.options.Pool.Go(func() {
		defer b.wg.Done()
		defer close(errC)
		err := b.Exec(ctx, cmd)
		if err != nil {
			errC <- err
			return
		}
	})
	if err == nil {
		return errC
	}
	goErrC := make(chan error, 1)
	goErrC <- fmt.Errorf("failed to go, %w", err)
	close(goErrC)
	return chanx.Combine(goErrC, errC)
}

func (b *bus) AsyncQuery(ctx context.Context, q any) (<-chan any, <-chan error) {
	resC := make(chan any, 1)
	errC := make(chan error, 1)
	b.wg.Add(1)
	err := b.options.Pool.Go(func() {
		defer b.wg.Done()
		defer close(resC)
		defer close(errC)
		res, err := b.Query(ctx, q)
		if err != nil {
			errC <- err
			return
		}
		resC <- res
	})
	if err == nil {
		return resC, errC
	}
	goErrC := make(chan error, 1)
	goErrC <- fmt.Errorf("failed to go, %w", err)
	close(goErrC)
	return resC, chanx.Combine(goErrC, errC)
}

func (b *bus) Close(ctx context.Context) error {
	b.inShutdown.Store(true)
	c := syncx.WaitNotify(&b.wg)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c:
			return nil
		}
	}
}

func (b *bus) shuttingDown() bool {
	return b.inShutdown.Load()
}

type handlerInfo struct {
	handlerVal    reflect.Value
	handlerMethod reflect.Method
	inType        reflect.Type
}

type option struct {
	Pool gopher.Gopher
}

func newOption(opts ...Option) *option {
	o := &option{}
	for _, opt := range opts {
		opt(o)
	}
	if o.Pool == nil {
		o.Pool = sample.Gopher{}
	}
	return o
}

type Option func(*option)

func Pool(pool gopher.Gopher) Option {
	return func(o *option) {
		o.Pool = pool
	}
}

func NewBus(opts ...Option) Bus {
	return &bus{
		commands:   sync.Map{},
		queries:    sync.Map{},
		wg:         sync.WaitGroup{},
		inShutdown: atomic.Bool{},
		options:    newOption(opts...),
	}
}
