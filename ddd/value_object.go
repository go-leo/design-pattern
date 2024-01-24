package ddd

// ValueObject as described in the DDD book.
// Value objects compare by the values of their attributes, they don't have an identity.
type ValueObject[T any] interface {

	// SameValueAs return ture if the given value object's and this value object's attributes are the same.
	SameValueAs(other T) bool

	// Copy return A safe, deep copy of this value object.
	Copy() T
}
