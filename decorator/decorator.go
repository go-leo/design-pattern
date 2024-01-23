package decorator

// Decorator allows us to write something like decorators to object.
// It can execute something before invoke or after.
type Decorator[T any] interface {
	// Decorate wraps the underlying obj, adding some functionality.
	Decorate(obj T) T
}

// The DecoratorFunc type is an adapter to allow the use of ordinary functions as Decorator.
type DecoratorFunc[T any] func(obj T) T

// Decorate call f(obj).
func (f DecoratorFunc[T]) Decorate(obj T) T {
	return f(obj)
}

// Chain decorates the given object with all middlewares.
func Chain[T any](obj T, middlewares ...Decorator[T]) T {
	for i := len(middlewares) - 1; i >= 0; i-- {
		obj = middlewares[i].Decorate(obj)
	}
	return obj
}
