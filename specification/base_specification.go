package specification

import "context"

type base[T any] struct {
	Predicate func(ctx context.Context, t T) bool
}

func (spec *base[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.Predicate(ctx, t)
}

func (spec *base[T]) And(another Specification[T]) Specification[T] {
	return And[T](spec, another)
}

func (spec *base[T]) Or(another Specification[T]) Specification[T] {
	return Or[T](spec, another)
}

func (spec *base[T]) Not(another Specification[T]) Specification[T] {
	return Not[T](another)
}

func (spec *base[T]) Conjunction(others ...Specification[T]) Specification[T] {
	return Conjunction[T](others...)
}

func (spec *base[T]) Disjunction(others ...Specification[T]) Specification[T] {
	return Disjunction[T](others...)
}
