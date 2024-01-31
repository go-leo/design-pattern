package ddd

type Aggregate[T any, ID any] interface {
	Root() Entity[T, ID]
}

type AggregateRoot interface {
}
