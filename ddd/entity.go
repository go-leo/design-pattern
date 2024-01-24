package ddd

// Entity as explained in the DDD book.
// Entities compare by identity, not by attributes.
type Entity[T any, ID any] interface {

	// SameIdentityAs return true if the identities are the same, regardless of other attributes.
	SameIdentityAs(other T) bool

	// Identity return the identity of this entity.
	Identity() ID
}
