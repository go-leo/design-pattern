package specification

import "context"

// Not used to create a new specification that is the inverse (NOT) of the given Spec.
type not[T any] struct {
	specification[T]
	Spec Specification[T]
}

func (spec *not[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return !spec.Spec.IsSatisfiedBy(ctx, t)
}
