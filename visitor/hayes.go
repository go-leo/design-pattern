package main

// Hayes implements its accept method.
type Hayes struct {
}

// Accept visitor.
func (receiver Hayes) Accept(visitor ModemVisitor) {
	visitor.Visit(receiver)
}
