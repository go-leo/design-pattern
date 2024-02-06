package trampoline

// Trampoline pattern allows to define recursive algorithms by iterative loop.
//
// When get is called on the returned Trampoline, internally it will iterate calling ‘jump’
// on the returned Trampoline as long as the concrete instance returned is more(Trampoline),
// stopping once the returned instance is done(Object).
//
// Essential we convert looping via recursion into iteration,
// the key enabling mechanism is the fact that more(Trampoline) is a lazy operation.
//
// T is type for returning result.
type Trampoline[T any] interface {
	Get() T

	// Jump to next stage.
	Jump() Trampoline[T]

	Result() T

	// Complete checks if complete.
	Complete() bool

	Done() Trampoline[T]
}

func More() {

}
