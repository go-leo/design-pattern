package eventsource

import (
	"github.com/go-leo/design-pattern/event"
	"time"
)

type StoredEvent struct {
	EventBody  string
	EventId    int64
	OccurredOn time.Time
	TypeName   string
}

func (e StoredEvent) toDomainEvent() event.Event {
	return nil
}
