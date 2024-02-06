package main

import "fmt"

type FishingBoat struct{}

func (receiver FishingBoat) Sail() {
	fmt.Println("The fishing boat is sailing")
}
