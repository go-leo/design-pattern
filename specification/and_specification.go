package specification

import "context"

// and used to create a new base that is the AND of two other specifications.
type and[T any] struct {
	base[T]
	Left  Specification[T]
	Right Specification[T]
}

func (spec *and[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return spec.Left.IsSatisfiedBy(ctx, t) && spec.Right.IsSatisfiedBy(ctx, t)
}
