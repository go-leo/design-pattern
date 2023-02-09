package specification

// AndSpecification used to create a new specification that is the AND of two other specifications.
type AndSpecification[T any] struct {
	BaseSpecification[T]
	left  Specification[T]
	right Specification[T]
}

func And[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &AndSpecification[T]{left: left, right: right}
}

func (spec *AndSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.left.IsSatisfiedBy(t) && spec.right.IsSatisfiedBy(t)
}
