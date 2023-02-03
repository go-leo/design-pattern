package ddd

type IHandlingEvent DomainEvent[HandlingEvent]

type HandlingEvent struct {
	ID string
}

func (h HandlingEvent) SameEventAs(other HandlingEvent) bool {
	return h.ID == other.ID
}
