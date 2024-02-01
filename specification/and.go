package specification

import "context"

// and used to create a new specification that is the AND of two other specifications.
type and[T any] struct {
	specification[T]
	Left  Specification[T]
	Right Specification[T]
}

func (spec *and[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.Left.IsSatisfiedBy(ctx, t) && spec.Right.IsSatisfiedBy(ctx, t)
}
