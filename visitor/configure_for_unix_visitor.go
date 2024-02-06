package main

import (
	"fmt"
)

type ConfigureForUnixVisitor struct{}

func (receiver ConfigureForUnixVisitor) Visit(modem Modem) {
	switch modem := modem.(type) {
	case Hayes:
		fmt.Printf("%T used with Unix configurator.\n", modem)
	case Zoom:
		fmt.Printf("%T used with Unix configurator.\n", modem)
	default:
		panic(fmt.Errorf("%T", modem))
	}
}
