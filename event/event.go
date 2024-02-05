package event

//
//import (
//	"context"
//	"time"
//)
//
//// Event is event.
//type Event[Body any, ID comparable] struct {
//	body       Body
//	id         ID
//	occurredOn time.Time
//	ctx        context.Context
//}
//
//// Body return the body of the Event
//func (e *Event[Body, ID]) Body() Body {
//	return e.body
//}
//
//// When return the time of the Event.
//func (e *Event[Body, ID]) When() time.Time {
//	return e.occurredOn
//}
//
//// ID return the id of the Event.
//func (e *Event[Body, ID]) ID() ID {
//	return e.id
//}
//
//// WithContext returns a shallow copy of e with its context changed to ctx.
//// The provided ctx must be non-nil.
//func (e *Event[Body, ID]) WithContext(ctx context.Context) *Event[Body, ID] {
//	if ctx == nil {
//		panic("nil context")
//	}
//	copied := new(Event[Body, ID])
//	*copied = *e
//	copied.ctx = ctx
//	return copied
//}
//
//// Context returns the context of the Event. To change the context, use WithContext.
//func (e *Event[Body, ID]) Context() context.Context {
//	if e.ctx != nil {
//		return e.ctx
//	}
//	return context.Background()
//}
