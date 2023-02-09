package specification

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

type BaseSpecification[T any] struct {
	isSatisfiedBy func(t T) bool
}

func New[T any](isSatisfiedBy func(t T) bool) Specification[T] {
	return &BaseSpecification[T]{isSatisfiedBy: isSatisfiedBy}
}

func (spec *BaseSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.isSatisfiedBy(t)
}

func (spec *BaseSpecification[T]) And(another Specification[T]) Specification[T] {
	return And[T](spec, another)
}

func (spec *BaseSpecification[T]) Or(another Specification[T]) Specification[T] {
	return Or[T](spec, another)
}

func (spec *BaseSpecification[T]) Not(another Specification[T]) Specification[T] {
	return Not[T](another)
}
