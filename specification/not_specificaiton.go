package specification

import "context"

// Not used to create a new base that is the inverse (NOT) of the given Spec.
type not[T any] struct {
	base[T]
	Spec Specification[T]
}

func (spec *not[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	return !spec.Spec.IsSatisfiedBy(ctx, t)
}
