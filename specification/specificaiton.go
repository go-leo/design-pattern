package specification

import "context"

// Specification interface.
// Use base as base for creating specifications, and
// only the method predicate(Object) must be implemented.
type Specification[T any] interface {

	// IsSatisfiedBy check if t is satisfied by the base.
	IsSatisfiedBy(ctx context.Context, t T) bool

	// And create a new base that is the AND operation of the current base and
	// another base.
	And(another Specification[T]) Specification[T]

	// Or create a new base that is the OR operation of the current base and
	// another base.
	Or(another Specification[T]) Specification[T]

	// Not create a new base that is the NOT operation of the current base.
	Not(another Specification[T]) Specification[T]

	// Conjunction create a new base that is conjunction operation of the current base.
	Conjunction(others ...Specification[T]) Specification[T]

	// Disjunction create a new base that is disjunction operation of the current base.
	Disjunction(others ...Specification[T]) Specification[T]
}

func New[T any](predicate func(ctx context.Context, t T) bool) Specification[T] {
	return &base[T]{Predicate: predicate}
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
