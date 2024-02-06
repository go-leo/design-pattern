package event

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-leo/gox/errorx"
	"github.com/go-leo/gox/syncx"
	"github.com/go-leo/gox/syncx/chanx"
	"reflect"
	"sync"
	"sync/atomic"
)

type Bus interface {
	// On adds a Listener to the event bus.
	On(listener any) error

	// Prepend adds the Listener to the beginning of the Listeners array.
	Prepend(listener any) error

	// Once adds a one-time Listener to the event bus.
	Once(listener any) error

	// PrependOnce adds a one-time Listener to the beginning of the Listeners array.
	PrependOnce(listener any) error

	// Emit synchronously calls each of the Listeners registered for the specified event,
	// in the order they were registered, passing the supplied arguments to each.
	Emit(event any) error

	// AsyncEmit asynchronously calls each of the Listeners registered.
	AsyncEmit(event any) <-chan error

	// Off removes the specified Listener from the Listeners array.
	Off(listener any) error

	// OffAll removes all Listeners, or those of the specified event.
	OffAll(event any) error

	// Close bus gracefully.
	Close(ctx context.Context) error
}

var _ Bus = (*bus)(nil)

type bus struct {
	listenerMap     sync.Map
	onceListenerMap sync.Map
	wg              sync.WaitGroup
	inShutdown      atomic.Bool // true when bus is in shutdown
	options         *option
}

func (b *bus) On(listener any) error {
	info, err := b.reflectListener(listener)
	if err != nil {
		return err
	}
	actual, loaded := b.listenerMap.LoadOrStore(info.eventType, wrapReflectListenerSlice([]*reflectListener{info}))
	if !loaded {
		return nil
	}
	listenerSlice := actual.(*syncReflectListenerSlice)
	listenerSlice = listenerSlice.Append(info)
	b.listenerMap.Store(info.eventType, listenerSlice)
	return nil
}

func (b *bus) Prepend(listener any) error {
	info, err := b.reflectListener(listener)
	if err != nil {
		return err
	}
	actual, loaded := b.listenerMap.LoadOrStore(info.eventType, wrapReflectListenerSlice([]*reflectListener{info}))
	if !loaded {
		return nil
	}
	listeners := actual.(*syncReflectListenerSlice)
	listeners = listeners.Prepend(info)
	b.listenerMap.Store(info.eventType, listeners)
	return nil
}

func (b *bus) Once(listener any) error {
	info, err := b.reflectListener(listener)
	if err != nil {
		return err
	}
	lis := &onceListener{Listener: info}
	actual, loaded := b.onceListenerMap.LoadOrStore(
		info.eventType,
		wrapOnceListenerSlice([]*onceListener{lis}))
	if !loaded {
		return nil
	}
	listeners := actual.(*syncOnceListenerSlice)
	listeners = listeners.Append(lis)
	b.onceListenerMap.Store(info.eventType, listeners)
	return nil
}

func (b *bus) PrependOnce(listener any) error {
	info, err := b.reflectListener(listener)
	if err != nil {
		return err
	}
	lis := &onceListener{Listener: info}
	actual, loaded := b.onceListenerMap.LoadOrStore(
		info.eventType,
		wrapOnceListenerSlice([]*onceListener{lis}))
	if !loaded {
		return nil
	}
	listeners := actual.(*syncOnceListenerSlice)
	listeners = listeners.Prepend(lis)
	b.onceListenerMap.Store(info.eventType, listeners)
	return nil
}

func (b *bus) Emit(event any) error {
	if event == nil {
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

func (b *bus) AsyncEmit(event any) <-chan error {
	//TODO implement me
	panic("implement me")
}

func (b *bus) Off(listener any) error {
	//TODO implement me
	panic("implement me")
}

func (b *bus) OffAll(event any) error {
	//TODO implement me
	panic("implement me")
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

func (b *bus) reflectListener(listener any) (*reflectListener, error) {
	if listener == nil {
		return nil, ErrHandlerNil
	}
	if b.shuttingDown() {
		return nil, ErrBusClosed
	}
	listenerVal := reflect.ValueOf(listener)
	listenerType := listenerVal.Type()
	method, ok := listenerType.MethodByName("Handle")
	if !ok {
		return nil, ErrUnimplemented
	}
	if method.Type.NumIn() != 2 {
		return nil, ErrUnimplemented
	}
	if method.Type.NumOut() != 1 {
		return nil, ErrUnimplemented
	}
	if !method.Type.Out(0).Implements(errorx.ErrorType) {
		return nil, ErrUnimplemented
	}
	eventType := method.Type.In(1)
	return &reflectListener{
		listenerVal:    listenerVal,
		listenerMethod: method,
		eventType:      eventType,
	}, nil
}

func (b *bus) Exec(cmd any) error {
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
	info := value.(*listenerInfo)
	resultValues := info.listenerMethod.Func.Call(
		[]reflect.Value{
			info.listenerVal,
			reflect.ValueOf(cmd),
		})
	err := resultValues[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
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
