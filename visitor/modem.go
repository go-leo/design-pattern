package main

type Modem interface {
	Accept(visitor ModemVisitor)
}
