package event

import (
	"context"
	"errors"
	"github.com/go-leo/gox/slicex"
	"github.com/go-leo/gox/syncx"
	"github.com/go-leo/gox/syncx/chanx"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

type Bus interface {
	// On adds a Listener to the Event bus.
	On(e Event, lis Listener) error

	// Prepend adds the Listener to the beginning of the listeners.
	Prepend(e Event, lis Listener) error

	// Once adds a one-time Listener to the Event bus.
	Once(e Event, lis Listener) error

	// PrependOnce adds a one-time Listener to the beginning of the listeners.
	PrependOnce(e Event, lis Listener) error

	// Emit synchronously calls each of the listeners registered for the specified Event,
	// in the order they were registered.
	Emit(e Event) error

	// AsyncEmit asynchronously calls each of the listeners registered for the specified Event.
	AsyncEmit(e Event) <-chan error

	// Off removes the specified Listener from the listeners.
	Off(e Event, lis Listener) error

	// OffAll removes all listeners for the specified Event.
	OffAll(e Event) error

	// SetMaxListeners increases the max listeners of the bus.
	// SetMaxListeners(n int)

	// GetMaxListeners returns the current max listener value for the bus.
	// GetMaxListeners() int

	// Listeners returns a copy of the listeners for the event.
	// Listeners(e Event) []Listener

	// RawListeners returns a copy of listeners and wrappers for the event.
	// RawListeners(e Event) []Listener

	// ListenerCount returns the number of listeners listening to the event.
	// ListenerCount(e Event) int

	// Events returns an slice listing the events for which the bus has registered listeners.
	// Events() []Event

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

func (b *bus) On(e Event, lis Listener) error {
	if err := b.check(e, lis); err != nil {
		return err
	}
	b.spin(&b.listenerMap, e.Type(), lis, b.appendListener)
	return nil
}

func (b *bus) Prepend(e Event, lis Listener) error {
	if err := b.check(e, lis); err != nil {
		return err
	}
	b.spin(&b.listenerMap, e.Type(), lis, b.prependListener)
	return nil
}

func (b *bus) Once(e Event, lis Listener) error {
	if err := b.check(e, lis); err != nil {
		return err
	}
	onceLis := &onceListener{Listener: lis, Once: sync.Once{}}
	b.spin(&b.onceListenerMap, e.Type(), onceLis, b.appendListener)
	return nil
}

func (b *bus) PrependOnce(e Event, lis Listener) error {
	if err := b.check(e, lis); err != nil {
		return err
	}
	onceLis := &onceListener{Listener: lis, Once: sync.Once{}}
	b.spin(&b.onceListenerMap, e.Type(), onceLis, b.prependListener)
	return nil
}

func (b *bus) Emit(e Event) error {
	if err := b.checkEvent(e); err != nil {
		return ErrEventNil
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	value, ok := b.listenerMap.Load(e.Type())
	if !ok {
		return nil
	}
	listeners := value.(*[]Listener)
	errs := make([]error, 0, len(*listeners))
	for _, listener := range *listeners {
		errs = append(errs, listener.Handle(e))
	}
	return errors.Join(errs...)
}

func (b *bus) AsyncEmit(e Event) <-chan error {
	if err := b.checkEvent(e); err != nil {
		errC := make(chan error, 1)
		errC <- err
		close(errC)
		return errC
	}
	if b.shuttingDown() {
		errC := make(chan error, 1)
		errC <- ErrBusClosed
		close(errC)
		return errC
	}
	eventType := e.Type()
	value, ok := b.listenerMap.Load(eventType)
	if !ok {
		errC := make(chan error, 1)
		errC <- ErrListener{EventType: eventType}
		close(errC)
		return errC
	}
	listeners := value.(*[]Listener)
	errCs := make([]<-chan error, 0, len(*listeners))
	for _, listener := range *listeners {
		errC := make(chan error, 1)
		b.wg.Add(1)
		err := b.options.Pool.Go(func() {
			defer b.wg.Done()
			defer close(errC)
			err := listener.Handle(e)
			if err != nil {
				errC <- err
				return
			}
		})
		if err != nil {
			errC <- err
		}
		errCs = append(errCs, errC)
	}
	return chanx.Combine[error](errCs...)
}

func (b *bus) Off(e Event, lis Listener) error {
	if err := b.check(e, lis); err != nil {
		return err
	}
	eventType := e.Type()
	b.spin(&b.listenerMap, eventType, lis, b.offListener)
	b.spin(&b.onceListenerMap, eventType, lis, b.offOnceListener)
	return nil
}

func (b *bus) OffAll(e Event) error {
	if e == nil {
		return ErrEventNil
	}
	eventType := e.Type()
	if eventType.Kind() == reflect.Invalid {
		return ErrEventTypeInvalid
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	b.listenerMap.Delete(eventType)
	b.onceListenerMap.Delete(eventType)
	return nil
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

func (b *bus) check(e Event, lis Listener) error {
	if err := b.checkEvent(e); err != nil {
		return err
	}
	if err := b.checkListener(lis); err != nil {
		return err
	}
	if b.shuttingDown() {
		return ErrBusClosed
	}
	return nil
}

func (b *bus) checkEvent(e Event) error {
	if e == nil {
		return ErrEventNil
	}
	if e.Type().Kind() == reflect.Invalid {
		return ErrEventTypeInvalid
	}
	return nil
}

func (b *bus) checkListener(lis Listener) error {
	if lis == nil {
		return ErrListenerNil
	}
	if !reflect.TypeOf(lis).Comparable() {
		return ErrListenerIncomparable
	}
	return nil
}

func (*bus) loadAndOn(listenerMap *sync.Map, eventType reflect.Type, lis Listener, pendFunc func([]Listener, ...Listener) []Listener) (any, any, bool) {
	ptr := &[]Listener{lis}
	oldVal, ok := listenerMap.LoadOrStore(eventType, ptr)
	if !ok {
		return oldVal, nil, false
	}
	newListeners := pendFunc(*ptr, *(oldVal.(*[]Listener))...)
	newVal := &newListeners
	return oldVal, newVal, true
}

func (b *bus) appendListener(listenerMap *sync.Map, eventType reflect.Type, lis Listener) (any, any, bool) {
	pendFunc := func(listeners []Listener, listener ...Listener) []Listener {
		return append(listeners, listener...)
	}
	return b.loadAndOn(listenerMap, eventType, lis, pendFunc)
}

func (b *bus) prependListener(listenerMap *sync.Map, eventType reflect.Type, lis Listener) (any, any, bool) {
	pendFunc := func(listeners []Listener, listener ...Listener) []Listener {
		return slicex.Prepend(listeners, listener...)
	}
	return b.loadAndOn(listenerMap, eventType, lis, pendFunc)
}

func (*bus) loadAndOff(listenerMap *sync.Map, eventType reflect.Type, lis Listener, indexesFunc func([]Listener, Listener) []int) (any, any, bool) {
	oldVal, ok := listenerMap.Load(eventType)
	if !ok {
		return oldVal, nil, false
	}
	oldPtr := oldVal.(*[]Listener)
	if len(*oldPtr) == 0 {
		return oldVal, nil, false
	}
	indexes := indexesFunc(*oldPtr, lis)
	if len(indexes) <= 0 {
		return oldVal, nil, false
	}
	newListeners := slicex.DeleteAll(*oldPtr, indexes...)
	newVal := &newListeners
	return oldVal, newVal, true
}

func (b *bus) offListener(listenerMap *sync.Map, eventType reflect.Type, lis Listener) (any, any, bool) {
	indexesFunc := func(listeners []Listener, lis Listener) []int {
		return slicex.Indexes(listeners, lis)
	}
	return b.loadAndOff(listenerMap, eventType, lis, indexesFunc)
}

// offOnceListener 删除once
func (b *bus) offOnceListener(listenerMap *sync.Map, eventType reflect.Type, lis Listener) (any, any, bool) {
	indexesFunc := func(listeners []Listener, lis Listener) []int {
		f := func(onceLis Listener) bool {
			if onceLis.(*onceListener).Listener == lis {
				return true
			}
			return false
		}
		return slicex.IndexesFunc(listeners, f)
	}
	return b.loadAndOff(listenerMap, eventType, lis, indexesFunc)
}

func (b *bus) spin(listenerMap *sync.Map, eventType reflect.Type, lis Listener, load func(listenerMap *sync.Map, eventType reflect.Type, lis Listener) (any, any, bool)) {
	var oldVal any
	var newVal any
	var ok bool
	oldVal, newVal, ok = load(listenerMap, eventType, lis)
	if !ok {
		return
	}
	backoff := 1
	for !listenerMap.CompareAndSwap(eventType, oldVal, &newVal) {
		// Leverage the exponential backoff algorithm, see https://en.wikipedia.org/wiki/Exponential_backoff.
		for i := 0; i < backoff; i++ {
			runtime.Gosched()
		}
		if backoff < b.options.MaxBackoff {
			backoff <<= 1
		}
		oldVal, newVal, ok = load(listenerMap, eventType, lis)
		if !ok {
			return
		}
	}
}
