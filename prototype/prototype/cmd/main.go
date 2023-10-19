package main

import (
	"fmt"
	"github.com/go-leo/design-pattern/prototype/prototype"
)

func main() {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}
	_, err := prototype.Marshal(group)
	if err != nil {
		fmt.Println("error:", err)
	}
}
