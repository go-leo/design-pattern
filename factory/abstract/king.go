package main

// King interface.
type King interface {
	GetDescription() string
}

// ElfElfKing This is the elven castle.
type ElfElfKing struct{}

func (ElfElfKing) GetDescription() string {
	return "This is the elven king!"
}

// OrcKing This is the orc king.
type OrcKing struct{}

func (OrcKing) GetDescription() string {
	return "This is the orc king!"
}
