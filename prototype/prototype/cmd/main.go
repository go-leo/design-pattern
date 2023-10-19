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

	type Toy struct {
		Model      string
		ColorGroup ColorGroup
		Materials  map[string]string
	}

	toy := Toy{
		Model: "bumblebee",
		ColorGroup: ColorGroup{
			ID:     1,
			Name:   "Reds",
			Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
		},
		Materials: map[string]string{
			"head": "iron",
			"body": "plastic",
		},
	}

	_, err := prototype.Marshal(toy)
	if err != nil {
		fmt.Println("error:", err)
	}

}
