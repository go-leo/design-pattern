package main

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
