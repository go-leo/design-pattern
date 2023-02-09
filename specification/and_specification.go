package specification

// andSpecification used to create a new specification that is the AND of two other specifications.
type andSpecification[T any] struct {
	specification[T]
	left  Specification[T]
	right Specification[T]
}

func And[T any](left Specification[T], right Specification[T]) Specification[T] {
	return &andSpecification[T]{left: left, right: right}
}

func (spec *andSpecification[T]) IsSatisfiedBy(t T) bool {
	return spec.left.IsSatisfiedBy(t) && spec.right.IsSatisfiedBy(t)
}
