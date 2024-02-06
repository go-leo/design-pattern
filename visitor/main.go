package main

// The Acyclic Visitor pattern allows new functions to be added to existing class hierarchies
// without affecting those hierarchies, and without creating the dependency cycles that are inherent
// to the GoF Visitor pattern, by making the Visitor base class degenerate
//
// In this example the visitor base class is ModemVisitor. The base class of the visited
// hierarchy is Modem and has two children Hayes and Zoom each one having
// its own visitor interface HayesVisitor and ZoomVisitor respectively. ConfigureForUnixVisitor and
// ConfigureForDosVisitor implement each derivative's visit method only if it is required
func main() {
	conDos := ConfigureForDosVisitor{}
	conUnix := ConfigureForUnixVisitor{}

	zoom := Zoom{}
	hayes := Hayes{}

	hayes.Accept(conDos)  // Hayes modem with Dos configurator
	zoom.Accept(conDos)   // Zoom modem with Dos configurator
	hayes.Accept(conUnix) // Hayes modem with Unix configurator
	zoom.Accept(conUnix)  // Zoom modem with Unix configurator
}
