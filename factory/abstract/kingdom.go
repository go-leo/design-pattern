package main

import "errors"

// Kingdom Helper struct to manufacture KingdomFactory.
type Kingdom struct {
	Army   Army
	Castle Castle
	King   King
}

type KingdomType int

const (
	ElfKingdomType = iota
	OrcKingdomType
)

// FactoryMaker The factory of kingdom factories.
type FactoryMaker struct{}

func (FactoryMaker) MakeFactory(t KingdomType) KingdomFactory {
	switch t {
	case ElfKingdomType:
		return &ElfKingdomFactory{}
	case OrcKingdomType:
		return OrcKingdomFactory{}
	default:
		panic(errors.New("kingdom type not supported"))
	}
}
