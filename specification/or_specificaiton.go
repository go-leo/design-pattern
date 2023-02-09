package specification

import "context"

// OrSpecification used to create a new specification that is the OR of two other specifications.
type OrSpecification[T any] struct {
	BaseSpecification[T]
	left  Specification[T]
	right Specification[T]
}

func Or[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &OrSpecification[T]{left: left, right: right}
}

func (spec *OrSpecification[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.left.IsSatisfiedBy(ctx, t) || spec.right.IsSatisfiedBy(ctx, t)
}
