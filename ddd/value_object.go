package ddd

// ValueObject as described in the DDD book.
// Value objects compare by the values of their attributes, they don't have an identity.
type ValueObject[T any] interface {
	SameValueAs(other T) bool
}
