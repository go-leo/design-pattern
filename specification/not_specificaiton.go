package specification

// notSpecification used to create a new specification that is the inverse (NOT) of the given spec.
type notSpecification[T any] struct {
	specification[T]
	spec Specification[T]
}

func Not[T any](spec Specification[T]) Specification[T] {
	return &notSpecification[T]{spec: spec}
}

func (spec *notSpecification[T]) IsSatisfiedBy(t T) bool {
	return !spec.spec.IsSatisfiedBy(t)
}
