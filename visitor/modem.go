package main

// Modem interface.
type Modem interface {
	Accept(visitor ModemVisitor)
}
