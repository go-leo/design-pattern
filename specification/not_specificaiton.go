package specification

// NotSpecification used to create a new specification that is the inverse (NOT) of the given spec.
type NotSpecification[T any] struct {
	BaseSpecification[T]
	spec Specification[T]
}

func Not[T any](spec Specification[T]) Specification[T] {
	return &NotSpecification[T]{spec: spec}
}

func (spec *NotSpecification[T]) IsSatisfiedBy(t T) bool {
	return !spec.spec.IsSatisfiedBy(t)
}
