package specification

import "context"

// conjunction used to create a new specification that is the OR of two other specifications.
type conjunction[T any] struct {
	specification[T]
	Specs []Specification[T]
}

func (spec *conjunction[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	for _, spec := range spec.Specs {
		if !spec.IsSatisfiedBy(ctx, t) {
			return false
		}
	}
	return true
}
