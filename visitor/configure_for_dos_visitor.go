package main

import "fmt"

// ConfigureForDosVisitor implements both zoom's and  hayes' visit method for Dos
// manufacturer.
type ConfigureForDosVisitor struct{}

func (receiver ConfigureForDosVisitor) Visit(modem Modem) {
	switch modem := modem.(type) {
	case Hayes:
		fmt.Printf("%T used with Dos configurator.\n", modem)
	case Zoom:
		fmt.Printf("%T used with Dos configurator.\n", modem)
	default:
		panic(fmt.Errorf("%T", modem))
	}
}
