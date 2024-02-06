package main

// Castle interface.
type Castle interface {
	GetDescription() string
}

// ElfCastle This is the elven castle.
type ElfCastle struct{}

func (ElfCastle) GetDescription() string {
	return "This is the elven castle!"
}

// OrcCastle This is the orc castle.
type OrcCastle struct{}

func (OrcCastle) GetDescription() string {
	return "This is the orc castle!"
}
