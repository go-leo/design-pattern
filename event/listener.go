package event

import "sync"

// Listener is Event listener interface.
type Listener interface {
	// Handle handles Event logic.
	Handle(event Event) error
}

type onceListener struct {
	Listener Listener
	Once     sync.Once
}

func (listener *onceListener) Handle(event Event) error {
	var err error
	listener.Once.Do(func() {
		err = listener.Listener.Handle(event)
	})
	return err
}

//var wrapReflectListenerSlice = syncx.WrapSlice[[]*reflectListener]
//
//type syncReflectListenerSlice = syncx.Slice[[]*reflectListener, *reflectListener]
//
//type reflectListener struct {
//	listenerVal    reflect.Value
//	listenerMethod reflect.Method
//	eventType      reflect.Type
//}
//
//func (listener *reflectListener) Handle(event Event) error {
//	resultValues := listener.listenerMethod.Func.Call(
//		[]reflect.Value{
//			listener.listenerVal,
//			reflect.ValueOf(event),
//		})
//	err := resultValues[0].Interface()
//	if err != nil {
//		return err.(error)
//	}
//	return nil
//}
//
//var wrapOnceListenerSlice = syncx.WrapSlice[[]*onceListener]
//
//type syncOnceListenerSlice = syncx.Slice[[]*onceListener, *onceListener]
