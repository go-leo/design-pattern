package specification

// orSpecification used to create a new specification that is the OR of two other specifications.
type orSpecification[T any] struct {
	specification[T]
	left  Specification[T]
	right Specification[T]
}

func Or[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &orSpecification[T]{left: left, right: right}
}

func (spec *orSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.left.IsSatisfiedBy(t) || spec.right.IsSatisfiedBy(t)
}
