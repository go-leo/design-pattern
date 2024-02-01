package specification

import "context"

// Specification interface.
// Use specification as specification for creating specifications, and
// only the method predicate(Object) must be implemented.
type Specification[T any] interface {

	// IsSatisfiedBy check if t is satisfied by the specification.
	IsSatisfiedBy(ctx context.Context, t T) bool

	// And create a new specification that is the AND operation of the current specification and
	// another specification.
	And(another Specification[T]) Specification[T]

	// Or create a new specification that is the OR operation of the current specification and
	// another specification.
	Or(another Specification[T]) Specification[T]

	// Not create a new specification that is the NOT operation of the current specification.
	Not(another Specification[T]) Specification[T]

	// Conjunction create a new specification that is conjunction operation of the current specification.
	Conjunction(others ...Specification[T]) Specification[T]

	// Disjunction create a new specification that is disjunction operation of the current specification.
	Disjunction(others ...Specification[T]) Specification[T]
}

type specification[T any] struct {
	Predicate func(ctx context.Context, t T) bool
}

func (spec *specification[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.Predicate(ctx, t)
}

func (spec *specification[T]) And(another Specification[T]) Specification[T] {
	return And[T](spec, another)
}

func (spec *specification[T]) Or(another Specification[T]) Specification[T] {
	return Or[T](spec, another)
}

func (spec *specification[T]) Not(another Specification[T]) Specification[T] {
	return Not[T](another)
}

func (spec *specification[T]) Conjunction(others ...Specification[T]) Specification[T] {
	return Conjunction[T](others...)
}

func (spec *specification[T]) Disjunction(others ...Specification[T]) Specification[T] {
	return Disjunction[T](others...)
}

func New[T any](predicate func(ctx context.Context, t T) bool) Specification[T] {
	return &specification[T]{Predicate: predicate}
}

func And[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &and[T]{Left: left, Right: right}
}

func Not[T any](spec Specification[T]) Specification[T] {
	return &not[T]{Spec: spec}
}

func Or[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &or[T]{Left: left, Right: right}
}

func Conjunction[T any](specs ...Specification[T]) Specification[T] {
	return &conjunction[T]{Specs: specs}
}

func Disjunction[T any](specs ...Specification[T]) Specification[T] {
	return &disjunction[T]{Specs: specs}
}
