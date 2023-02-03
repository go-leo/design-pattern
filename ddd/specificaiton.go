package ddd

// Specification interface.
// Use BaseSpecification as base for creating specifications, and
// only the method isSatisfiedBy(Object) must be implemented.
type Specification[T any] interface {
	// IsSatisfiedBy check if t is satisfied by the specification.
	IsSatisfiedBy(t T) bool
	// And create a new specification that is the AND operation of the current specification and
	// another specification.
	And(another Specification[T]) Specification[T]
	// Or create a new specification that is the OR operation of the current specification and
	// another specification.
	Or(another Specification[T]) Specification[T]
	// Not create a new specification that is the NOT operation of the current specification.
	Not(another Specification[T]) Specification[T]
}

type specification[T any] struct {
	isSatisfiedBy func(t T) bool
}

func NewSpecification[T any](isSatisfiedBy func(t T) bool) Specification[T] {
	return &specification[T]{isSatisfiedBy: isSatisfiedBy}
}

func (spec *specification[T]) IsSatisfiedBy(t T) bool {
	return spec.isSatisfiedBy(t)
}

func (spec *specification[T]) And(another Specification[T]) Specification[T] {
	return NewAndSpecification[T](spec, another)
}

func (spec *specification[T]) Or(another Specification[T]) Specification[T] {
	return NewOrSpecification[T](spec, another)
}

func (spec *specification[T]) Not(another Specification[T]) Specification[T] {
	return NewNotSpecification[T](another)
}

// AndSpecification used to create a new specification that is the AND of two other specifications.
type AndSpecification[T any] struct {
	specification[T]
	left  Specification[T]
	right Specification[T]
}

func NewAndSpecification[T any](left Specification[T], right Specification[T]) *AndSpecification[T] {
	return &AndSpecification[T]{left: left, right: right}
}

func (spec *AndSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.left.IsSatisfiedBy(t) && spec.right.IsSatisfiedBy(t)
}

// OrSpecification used to create a new specification that is the OR of two other specifications.
type OrSpecification[T any] struct {
	specification[T]
	left  Specification[T]
	right Specification[T]
}

func NewOrSpecification[T any](left Specification[T], right Specification[T]) *OrSpecification[T] {
	return &OrSpecification[T]{left: left, right: right}
}

func (spec *OrSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.left.IsSatisfiedBy(t) || spec.right.IsSatisfiedBy(t)
}

// NotSpecification used to create a new specification that is the inverse (NOT) of the given spec.
type NotSpecification[T any] struct {
	specification[T]
	spec Specification[T]
}

func NewNotSpecification[T any](spec Specification[T]) *NotSpecification[T] {
	return &NotSpecification[T]{spec: spec}
}

func (spec *NotSpecification[T]) IsSatisfiedBy(t T) bool {
	return !spec.spec.IsSatisfiedBy(t)
}
