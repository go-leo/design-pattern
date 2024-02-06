package main

// Army interface.
type Army interface {
	GetDescription() string
}

// ElfArmy This is the elven army.
type ElfArmy struct{}

func (ElfArmy) GetDescription() string {
	return "This is the elven army!"
}

// OrcArmy This is the orc army
type OrcArmy struct{}

func (OrcArmy) GetDescription() string {
	return "This is the orc army!"
}
