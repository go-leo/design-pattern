package main

type FishingBoatAdapter struct {
	FishingBoat FishingBoat
}

func (receiver FishingBoatAdapter) Row() {
	receiver.FishingBoat.Sail()
}
