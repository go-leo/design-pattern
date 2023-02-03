package ddd

import "time"

// DomainEvent is something that is unique, but does not have a lifecycle.
// The identity may be explicit, for example the sequence number of a payment,
// or it could be derived from various aspects of the event such as where, when and what
// has happened.
type DomainEvent[T any] interface {
	SameEventAs(other T) bool
}

type domainEvent[T any] struct {
	ID        T
	Version   string
	Timestamp time.Time
}
