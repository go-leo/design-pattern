package main

import "fmt"

func main() {
	maker := FactoryMaker{}
	elfFactory := maker.MakeFactory(ElfKingdomType)
	elfKingdom := Kingdom{
		Army:   elfFactory.CreateArmy(),
		Castle: elfFactory.CreateCastle(),
		King:   elfFactory.CreateKing(),
	}
	fmt.Println(elfKingdom.Army.GetDescription())
	fmt.Println(elfKingdom.Castle.GetDescription())
	fmt.Println(elfKingdom.King.GetDescription())

	orcFactory := maker.MakeFactory(OrcKingdomType)
	orcKingdom := Kingdom{
		Army:   orcFactory.CreateArmy(),
		Castle: orcFactory.CreateCastle(),
		King:   orcFactory.CreateKing(),
	}
	fmt.Println(orcKingdom.Army.GetDescription())
	fmt.Println(orcKingdom.Castle.GetDescription())
	fmt.Println(orcKingdom.King.GetDescription())
}
