package main

func main() {
	captain := &Captain{
		RowingBoat: &FishingBoatAdapter{
			FishingBoat: FishingBoat{},
		},
	}
	captain.Row()
}
