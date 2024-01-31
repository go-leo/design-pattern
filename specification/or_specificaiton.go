package specification

import "context"

// or used to create a new base that is the OR of two other specifications.
type or[T any] struct {
	base[T]
	Left  Specification[T]
	Right Specification[T]
}

func (spec *or[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.Left.IsSatisfiedBy(ctx, t) || spec.Right.IsSatisfiedBy(ctx, t)
}
