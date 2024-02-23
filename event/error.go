package event

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrListenerNil listener is nil
	ErrListenerNil = errors.New("listener is nil")

	// ErrEventNil event is nil
	ErrEventNil = errors.New("event is nil")

	// ErrBusClosed bus was closed
	ErrBusClosed = errors.New("bus was closed")

	// ErrEventTypeInvalid event type is invalid
	ErrEventTypeInvalid = errors.New("event type is invalid")

	// ErrListenerIncomparable listener is incomparable
	ErrListenerIncomparable = errors.New("listener is incomparable")
)

type ErrListener struct {
	EventType reflect.Type
}

func (e ErrListener) Error() string {
	return fmt.Sprintf("%s listener not found", e.EventType)
}
