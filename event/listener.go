package event

import (
	"github.com/go-leo/gox/syncx"
	"reflect"
	"sync"
)

// Listener is event listener interface.
type Listener[Event any] interface {
	// Handle handles event logic.
	Handle(event Event) error
}

var wrapReflectListenerSlice = syncx.WrapSlice[[]*reflectListener]

type syncReflectListenerSlice = syncx.Slice[[]*reflectListener, *reflectListener]

type reflectListener struct {
	listenerVal    reflect.Value
	listenerMethod reflect.Method
	eventType      reflect.Type
}

func (listener *reflectListener) Handle(event any) error {
	resultValues := listener.listenerMethod.Func.Call(
		[]reflect.Value{
			listener.listenerVal,
			reflect.ValueOf(event),
		})
	err := resultValues[0].Interface()
	if err != nil {
		return err.(error)
	}
	return nil
}

var wrapOnceListenerSlice = syncx.WrapSlice[[]*onceListener]

type syncOnceListenerSlice = syncx.Slice[[]*onceListener, *onceListener]

type onceListener struct {
	Listener *reflectListener
	Once     sync.Once
}

func (listener *onceListener) Handle(event any) error {
	var err error
	listener.Once.Do(func() {
		err = listener.Listener.Handle(event)
	})
	return err
}
