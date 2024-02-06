package main

type Hayes struct {
}

func (receiver Hayes) Accept(visitor ModemVisitor) {
	visitor.Visit(receiver)
}
