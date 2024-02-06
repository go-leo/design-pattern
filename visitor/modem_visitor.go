package main

type ModemVisitor interface {
	Visit(modem Modem)
}
