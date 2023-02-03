package ddd

// Entity as explained in the DDD book.
// Entities compare by identity, not by attributes.
type Entity[T any, ID any] interface {
	SameIdentityAs(other T) bool
	Identity() ID
}
