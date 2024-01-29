package decorator

// Decorator allows us to write something like decorators to component.
// It can execute something before invoke or after.
type Decorator[T any] interface {
	// Decorate wraps the underlying obj, adding some functionality.
	Decorate(component T) T
}

// Func is an adapter to allow the use of ordinary functions as Decorator.
type Func[T any] func(component T) T

// Decorate call f(obj).
func (f Func[T]) Decorate(component T) T {
	return f(component)
}

// Chain decorates the given object with all middlewares.
func Chain[T any](component T, decorators ...Decorator[T]) T {
	for i := len(decorators) - 1; i >= 0; i-- {
		component = decorators[i].Decorate(component)
	}
	return component
}
