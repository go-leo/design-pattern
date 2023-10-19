package prototype

// Marshaler is the interface implemented by types that
// can marshal themselves into valid JSON.
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}
