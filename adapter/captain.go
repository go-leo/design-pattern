package main

type Captain struct {
	RowingBoat RowingBoat
}

func (receiver Captain) Row() {
	receiver.RowingBoat.Row()
}
