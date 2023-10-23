package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-leo/design-pattern/prototype/prototype"
)

// func mainSample() {
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

	var clonedToy Toy
	err := prototype.Clone(&clonedToy, toy)
	if err != nil {
		fmt.Println("error:", err)
	}

}

func Unsupported() {
	type ColorGroup struct {
		ID      int
		Name    string
		Colors  []string
		Channel chan int
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

	err := prototype.Clone(nil, toy)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func JsonUnmarshal() {
	var b bool
	err := json.Unmarshal([]byte("true"), &b)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(b)
}
