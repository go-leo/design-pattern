package main

// Zoom implements its accept method.
type Zoom struct{}

// Accept visitor.
func (receiver Zoom) Accept(visitor ModemVisitor) {
	visitor.Visit(receiver)
}
