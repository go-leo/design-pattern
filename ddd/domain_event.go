package ddd

import (
	"encoding/json"
	"time"
)

// DomainEvent is something that is unique, but does not have a lifecycle.
// The identity may be explicit, for example the sequence number of a payment,
// or it could be derived from various aspects of the event such as where, when and what
// has happened.
type DomainEvent[T any, ID any] interface {
	SameEventAs(other T) bool
	ID() ID
	Kind() string
	When() time.Time
}

type EventDescriptor struct {
	ID         int64
	Body       string
	OccurredAt time.Time
	Kind       string
}

func NewEventDescriptor(body string, occurredAt time.Time, kind string) *EventDescriptor {
	return &EventDescriptor{Body: body, OccurredAt: occurredAt, Kind: kind}
}

func EventDescriptorFromEvent[T any, ID any](e DomainEvent[T, ID]) *EventDescriptor {
	data, _ := json.Marshal(e)
	return NewEventDescriptor(string(data), e.When(), e.Kind())
}

type domainEvent[T any] struct {
	ID        T
	Version   string
	Timestamp time.Time
}

///Users/songyancheng/Workspace/DDD/clean-arch-ddd-intro/src/main/java/com/github/felpexw/shared/domain/common
///Users/songyancheng/Workspace/DDD/event-source-cqrs-sample/src/main/java/io/dddbyexamples/eventsource/eventstore
///Users/songyancheng/Workspace/DDD/eventsourcing-java-example/eventsourcing/src/main/java/com/pragmatists/eventsourcing
///Users/songyancheng/Workspace/DDD/java-cqrs-intro/cqrswithes/src/main/java/pl/altkom/asc/lab/cqrs/intro/cqrswithes/cqs
///Users/songyancheng/Workspace/DDD/java-cqrs-intro/cqrswithes/src/main/java/pl/altkom/asc/lab/cqrs/intro/cqrswithes/db
