package main

type Zoom struct{}

func (receiver Zoom) Accept(visitor ModemVisitor) {
	visitor.Visit(receiver)
}
