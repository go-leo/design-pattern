package event

import (
	"context"
	"reflect"
	"time"
)

// Event is an interface with a specific type that is associated to a specific Listener.
type Event interface {

	// When return the time of the event.
	When() time.Time

	// ID return the id of the event.
	ID() any

	// Body return the body of the event.
	Body() any

	// Type return the body's reflect.Type of the event.
	Type() reflect.Type

	// WithContext returns a shallow copy of e with its context changed to ctx.
	// The provided ctx must be non-nil.
	WithContext(ctx context.Context) Event

	// Context returns the context of the event. To change the context, use WithContext.
	Context() context.Context
}

// event is event.
type event struct {
	body       any
	id         any
	occurredOn time.Time
	ctx        context.Context
}

// ID return the id of the event.
func (e *event) ID() any {
	return e.id
}

// When return the time of the event.
func (e *event) When() time.Time {
	return e.occurredOn
}

// Body return the body of the event
func (e *event) Body() any {
	return e.body
}

func (e *event) Type() reflect.Type {
	return reflect.TypeOf(e.body)
}

// WithContext returns a shallow copy of e with its context changed to ctx.
// The provided ctx must be non-nil.
func (e *event) WithContext(ctx context.Context) Event {
	if ctx == nil {
		panic("nil context")
	}
	copied := new(event)
	*copied = *e
	copied.ctx = ctx
	return copied
}

// Context returns the context of the event. To change the context, use WithContext.
func (e *event) Context() context.Context {
	if e.ctx != nil {
		return e.ctx
	}
	return context.Background()
}

func NewEvent(body any, id any) Event {
	return &event{body: body, id: id, occurredOn: time.Now()}
}
