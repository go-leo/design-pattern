package specification

import "context"

// disjunction used to create a new specification that is the OR of two other specifications.
type disjunction[T any] struct {
	specification[T]
	Specs []Specification[T]
}

func (spec *disjunction[T]) IsSatisfiedBy(ctx context.Context, t T) bool {
	for _, spec := range spec.Specs {
		if spec.IsSatisfiedBy(ctx, t) {
			return true
		}
	}
	return false
}
