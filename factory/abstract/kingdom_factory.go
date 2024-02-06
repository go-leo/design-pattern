package main

// KingdomFactory factory interface.
type KingdomFactory interface {
	CreateCastle() Castle

	CreateKing() King

	CreateArmy() Army
}

type ElfKingdomFactory struct{}

func (ElfKingdomFactory) CreateCastle() Castle {
	return &ElfCastle{}
}

func (ElfKingdomFactory) CreateKing() King {
	return &ElfElfKing{}
}

func (ElfKingdomFactory) CreateArmy() Army {
	return ElfArmy{}
}

type OrcKingdomFactory struct{}

func (OrcKingdomFactory) CreateCastle() Castle {
	return OrcCastle{}
}

func (OrcKingdomFactory) CreateKing() King {
	return OrcKing{}
}

func (OrcKingdomFactory) CreateArmy() Army {
	return OrcArmy{}
}
