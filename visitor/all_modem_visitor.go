package main

// AllModemVisitor interface extends all visitor interfaces. This interface provides ease of use
// when a visitor needs to visit all modem types.
type AllModemVisitor interface {
	HayesVisitor
	ZoomVisitor
}
