package main

import (
	"fmt"
)

// ConfigureForUnixVisitor implements zoom's visit method for Unix manufacturer, unlike
// traditional visitor pattern, this class may selectively implement visit for other modems.
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
